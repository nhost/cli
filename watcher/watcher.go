package watcher

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type GitWatcher struct {
	status        *util.Status
	log           logrus.FieldLogger
	repoExists    bool
	branch        string
	headRef       string
	remoteRef     string
	repoCreatedCh chan bool
	branchCh      chan string
	refCh         chan string
}

func NewGitWatcher(status *util.Status, log logrus.FieldLogger) *GitWatcher {
	return &GitWatcher{
		status:        status,
		log:           log,
		repoCreatedCh: make(chan bool),
		branchCh:      make(chan string),
		refCh:         make(chan string),
		repoExists:    util.PathExists(nhost.GIT_DIR),
	}
}

func (gw *GitWatcher) Watch(ctx context.Context, interval time.Duration, reloadFunc func(branch, ref string) error) {
	go gw.watchRepoExists(ctx, interval)
	go gw.watchBranchChange(ctx, interval)
	go gw.watchRefChange(ctx, interval)

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-gw.repoCreatedCh:
			gw.status.Infoln("Git repo created")
			if !ok {
				// nullify the channel, so it's not ready for communication and we don't fall into an infinite loop
				gw.repoCreatedCh = nil
			}
		case branch := <-gw.branchCh:
			gw.status.Infoln(fmt.Sprintf("Detected branch change: %s", branch))
			gw.log.WithField("branch", branch).Debug("Detected branch change")

			if err := reloadFunc(gw.branch, gw.remoteRef); err != nil {
				gw.log.WithError(err).Errorln("Failed to reload")
				gw.status.Errorln(fmt.Sprintf("Failed to reload: %v", err))
			}
		case ref := <-gw.refCh:
			gw.status.Infoln(fmt.Sprintf("Detected remoteRef change: %s", ref))
			gw.log.WithField("ref", ref[:7]).Debug("Detected git remote ref change")

			if err := reloadFunc(gw.branch, gw.remoteRef); err != nil {
				gw.log.WithError(err).Errorln("Failed to reload")
				gw.status.Errorln(fmt.Sprintf("Failed to reload: %v", err))
			}
		}
	}
}

func (gw *GitWatcher) watchBranchChange(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			if !util.PathExists(filepath.Join(nhost.GIT_DIR, "HEAD")) {
				continue
			}

			branch := nhost.GetCurrentBranch()
			if gw.branch == "" {
				// if the branch is not set yet, do not notify to prevent reload triggering
				gw.branch = branch
				continue
			}

			if gw.branch != branch {
				gw.branch = branch
				gw.branchCh <- branch
			}
		}
	}
}

func (gw *GitWatcher) watchRefChange(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			// TODO: the "origin" remote is hardcoded here, make it configurable
			if gw.branch == "" ||
				!util.PathExists(filepath.Join(nhost.GIT_DIR, "refs/remotes/origin", gw.branch)) ||
				!util.PathExists(filepath.Join(nhost.GIT_DIR, "refs/heads", gw.branch)) {
				continue
			}

			remoteBranchRef, err := nhost.GetRemoteBranchRef(gw.branch)
			if err != nil {
				gw.log.WithError(err).Errorln("Failed to get branch remoteRef")
				continue
			}

			headBranchRef, err := nhost.GetHeadBranchRef(gw.branch)
			if err != nil {
				gw.log.WithError(err).Errorln("Failed to get branch headRef")
				continue
			}

			if gw.remoteRef == "" || gw.headRef == "" {
				// if reds aren't set yet, do not notify to prevent reload triggering
				gw.remoteRef = remoteBranchRef
				gw.headRef = headBranchRef
				continue
			}

			if gw.headRef != headBranchRef && headBranchRef == remoteBranchRef {
				gw.headRef = headBranchRef
				gw.remoteRef = remoteBranchRef
				gw.refCh <- headBranchRef
			}
		}
	}
}

func (gw *GitWatcher) watchRepoExists(ctx context.Context, interval time.Duration) {
	if gw.repoExists {
		// if it's already there, we don't need to do anything
		return
	}

	ticker := time.NewTicker(interval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			if util.PathExists(nhost.GIT_DIR) {
				close(gw.repoCreatedCh)
				ticker.Stop()
				return
			}
		}
	}
}

type Operation func(event fsnotify.Event) error

type Watcher struct {
	log logrus.FieldLogger

	//	It's inherently an fsnotify Watcher under the hood.
	*fsnotify.Watcher

	//  In the following format:
	//  Key - Absolute File Name to Watch
	//  Value - Function to execute
	Map map[string]Operation

	//	(Optional) Context to use for stopping of watcher.
	context context.Context
}

//	Add individual location to watcher.
//	Along with associating it with respective operation function.
func (w *Watcher) Register(path string, op Operation) error {
	w.log.WithField("component", "path").Debugln("Watching", util.Rel(path))

	w.Map[path] = op

	return w.Add(path)
}

//	Validates whether a given key is already
//	register in the watcher.
func (w *Watcher) Registered(key string) bool {

	for name := range w.Map {
		if name == key {
			return true
		}
	}

	return false
}

//	Initializes a new watcher
func New(ctx context.Context, logger logrus.FieldLogger) (*Watcher, error) {

	//	If no context has been supplied,
	//	initialize a new one
	if ctx == nil {
		ctx = context.Background()
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize a watcher")
	}

	return &Watcher{
		log:     logger,
		context: ctx,
		Map:     make(map[string]Operation),
		Watcher: w,
	}, nil
}

//  Infinite function which listens for
//  fsnotify events once launched
func (w *Watcher) Start() {

	w.log.WithField("component", "watcher").Debug("Activated")

	for {
		select {

		//  Inactivate the watch when the environment shuts does
		case <-w.context.Done():
			w.log.WithField("component", "watcher").Debug("Inactivated")
			return

		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {

				//  run the operation
				go func() {
					if err := w.Map[event.Name](event); err != nil {
						w.log.WithField("component", "watcher").Debug(err)
					}
				}()
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			w.log.WithField("component", "watcher").Debug(err)
		}
	}
}

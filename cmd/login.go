/*
MIT License

# Copyright (c) Nhost

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	email    string
	password string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:        "login",
	SuggestFor: []string{"logout"},
	Short:      "Log in to your Nhost account",
	PreRun: func(cmd *cobra.Command, args []string) {
		//  if user is already logged in, ask to logout
		if _, err := getUser(nhost.AUTH_PATH); err == nil {
			status.Fatal(ErrLoggedIn)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if email == "" {
			readEmail, err := readInput("email", false)
			if err != nil {
				os.Exit(0)
			}
			email = readEmail
		}

		if password == "" {
			payload, err := readInput("password", true)
			if err != nil {
				os.Exit(0)
			}
			password = payload
		}

		status.Info("Authenticating")
		credentials, err := login(nhost.API, email, password)
		if err != nil {
			status.Error("Failed to login with that email")
			return err
		}

		//  delete any existing auth files
		if util.PathExists(nhost.AUTH_PATH) {
			if err = util.DeletePath(nhost.AUTH_PATH); err != nil {
				status.Error(fmt.Sprintf("Failed to reset the auth file, please delete it manually from: %s, and re-run `nhost login`", nhost.AUTH_PATH))
				return err
			}
		}

		//  create the auth file path if it doesn't exist
		err = os.MkdirAll(nhost.ROOT, os.ModePerm)
		if err != nil {
			status.Error("Failed to initialize Nhost root directory: " + nhost.ROOT)
			return err
		}

		//  create the auth file to write it
		f, err := os.Create(nhost.AUTH_PATH)
		if err != nil {
			status.Error("Failed to create auth configuration file")
			return err
		}

		defer f.Close()

		//  write auth file
		output, _ := json.Marshal(credentials)
		return writeToFile(nhost.AUTH_PATH, string(output), "end")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		status.Info("Email verified, and you are logged in!")
		status.Info("Type `nhost list` to see your remote apps")
	},
}

// take email input from user
func readInput(key string, hide bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var response string
	var err error

	fmt.Print(util.Bold + strings.Title(key) + ": " + util.Reset)
	if !hide {
		response, err = reader.ReadString('\n')
	} else {
		output, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return response, err
		}
		response = string(output)
	}

	return strings.TrimSpace(response), err
}

// Gets user's details using specified token
func getUser(authFile string) (nhost.User, error) {
	var response nhost.User
	if !util.PathExists(authFile) {
		return response, errors.New("auth source not found")
	}

	log.Debug("Fetching user data")

	credentials, err := nhost.LoadCredentials()
	if err != nil {
		return response, err
	}

	//	Encode the data
	postBody, _ := json.Marshal(credentials)
	responseBody := bytes.NewBuffer(postBody)

	// Leverage Go's HTTP Post function to make request
	resp, err := http.Post(nhost.API+"/custom/cli/user", "application/json", responseBody)
	if err != nil {
		return response, err
	}

	//  read our opened xmlFile as a byte array.
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	defer resp.Body.Close()

	if response.ID == "" {
		err = errors.New("user not found")
	}

	return response, err
}

// signs the user in with email and returns verification token
func login(url, email, password string) (nhost.Credentials, error) {
	log.Debug("Authenticating with ", email)

	var response nhost.Credentials

	//	Encode the data
	postBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	responseBody := bytes.NewBuffer(postBody)

	//	Leverage Go's HTTP Post function to make request
	resp, err := http.Post(url+"/custom/cli/login", "application/json", responseBody)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	//Handle Error
	/*
		if response.Error.Code == "not_found" {
			return response.VerificationToken, errors.New("we couldn't find an account registered with this email, please register at https://nhost.io/register")
		} else if response.Error.Code == "unknown" {
			return response.VerificationToken, errors.New("error while trying to create a login token")
		} else if response.Error.Code == "server_not_available" {
			return response.VerificationToken, errors.New("service unavailable")
		}
	*/

	return response, err
}

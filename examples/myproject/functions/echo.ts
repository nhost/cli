import { createGzip } from 'zlib'
import { IncomingMessage, ServerResponse } from 'http'
import process from 'process'

export default (req: IncomingMessage, res: ServerResponse) => {
    const responseData = {
        headers: req.headers,
        query: new URL(req.url || '', `http://${req.headers.host}`).searchParams,
        method: req.method,
        node: process.version,
        arch: process.arch,
    }

    const jsonString = JSON.stringify(responseData)
    const acceptsGzip = req.headers['accept-encoding']?.includes('gzip')

    if (acceptsGzip) {
        res.setHeader('Content-Encoding', 'gzip')
        res.setHeader('Content-Type', 'application/json')
        const gzip = createGzip()
        gzip.pipe(res)
        gzip.end(jsonString)
    } else {
        res.setHeader('Content-Type', 'application/json')
        res.end(jsonString)
    }
}

import { Request, Response } from 'express'
import process from 'process'
import zlib from 'zlib'

export default (req: Request, res: Response) => {
    var response = {
        headers: req.headers,
        query: req.query,
        node: process.version,
        arch: process.arch,
        data: {
            message: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.',
        },
    }

    // compress the response
    res.setHeader('Content-Encoding', 'gzip')
    res.setHeader('Content-Type', 'application/json')

    const compressedResponse = zlib.gzipSync(JSON.stringify(response))

    res.send(compressedResponse)
}

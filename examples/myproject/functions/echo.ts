import { Request, Response } from 'express'
import process from 'process'
import zlib from 'zlib'

export default (req: Request, res: Response) => {
    const response = {
        headers: req.headers,
        query: req.query,
        node: process.version,
        arch: process.arch,
        data: {
            message: 'Hello, World!',
        },
    }

    // compress the response
    res.setHeader('Content-Encoding', 'gzip')
    res.setHeader('Content-Type', 'application/json')

    const compressedResponse = zlib.gzipSync(JSON.stringify(response))

    res.send(compressedResponse)
}

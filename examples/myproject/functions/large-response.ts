import { createGzip } from 'zlib'
import { IncomingMessage, ServerResponse } from 'http'

export default (req: IncomingMessage, res: ServerResponse) => {
    const dataSize = 5 * 1024 * 1024
    const responseData = {
        size: dataSize,
        timestamp: new Date().toISOString(),
        data: Array(dataSize).fill('A').join('')
    }

    const jsonString = JSON.stringify(responseData)
    const acceptsGzip = req.headers['accept-encoding']?.includes('gzip')

    if (acceptsGzip) {
        res.setHeader('Content-Encoding', 'gzip')
        res.setHeader('Content-Type', 'application/json')
        const gzip = createGzip()
        gzip.pipe(res)
        gzip.write(jsonString)
        gzip.end()
    } else {
        res.setHeader('Content-Type', 'application/json')
        res.write(jsonString)
        res.end()
    }
} 
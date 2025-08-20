import { Request, Response } from 'express'
import { gzipSync } from 'zlib'

export default (req: Request, res: Response) => {
  const data = 'hello world'
  const compressed = gzipSync(Buffer.from(data, 'utf8'))
  const base64Compressed = compressed.toString('base64')

  res.setHeader('Content-Type', 'application/octet-stream')
  res.status(200).send(base64Compressed)
}

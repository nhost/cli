import { Request, Response } from 'express';
import zlib from 'zlib';

// Helper function to check if client accepts compression
function acceptsCompression(req: Request): boolean {
    const acceptEncoding = req.headers['accept-encoding'];
    return acceptEncoding !== undefined &&
           (acceptEncoding.includes('gzip') || acceptEncoding.includes('deflate'));
}

// Helper function to compress data
function compressResponse(data: any, req: Request): Promise<Buffer> {
    return new Promise((resolve, reject) => {
        const acceptEncoding = req.headers['accept-encoding'];
        const jsonString = JSON.stringify(data);

        if (acceptEncoding?.includes('gzip')) {
            zlib.gzip(jsonString, (error, result) => {
                if (error) reject(error);
                else resolve(result);
            });
        } else if (acceptEncoding?.includes('deflate')) {
            zlib.deflate(jsonString, (error, result) => {
                if (error) reject(error);
                else resolve(result);
            });
        } else {
            resolve(Buffer.from(jsonString));
        }
    });
}

export default async (req: Request, res: Response) => {
    const response = {
        headers: req.headers,
        query: req.query,
        node: process.version,
        arch: process.arch,
        data: {
            message: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.',
        },
    };

    if (acceptsCompression(req)) {
        try {
            const compressed = await compressResponse(response, req);
            const encoding = req.headers['accept-encoding']?.includes('gzip') ? 'gzip' : 'deflate';

            res.setHeader('Content-Encoding', encoding);
            res.setHeader('Content-Type', 'application/json');
            res.send(compressed);
        } catch (error) {
            // Fallback to uncompressed response if compression fails
            res.send(response);
        }
    } else {
        res.send(response);
    }
}

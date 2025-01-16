import { Request, Response } from 'express'
import process from 'process'

export default (req: Request, res: Response) => {
    try {
        throw new Error('This is an error')
    } catch (error) {
        console.log(error)
        res.status(502).json({
            error: error.message,
        })
    }
}


import { Request, Response } from 'express'

export default (_: Request, res: Response) => {
    try {
        throw new Error('This is an error')
    } catch (error) {
        console.log(error)

        new Promise(resolve => setTimeout(resolve, 1000))
            .then(() => {
                res.status(500).json({
                    error: error.message,
                })
            })
    }
}

import { Request, Response } from 'express'

export default async (req: Request, res: Response) => {
    try {
        const response = await fetch('https://api.ipify.org?format=json')
        const data = await response.json()
        
        res.status(200).json({
            ip: data.ip,
            timestamp: new Date().toISOString()
        })
    } catch (error) {
        res.status(500).json({
            error: 'Failed to fetch IP address',
            message: error instanceof Error ? error.message : 'Unknown error'
        })
    }
}
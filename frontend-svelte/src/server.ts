import sirv from 'sirv'
import polka from 'polka'
import compression from 'compression'
import * as sapper from '@sapper/server'
import { createProxyMiddleware } from 'http-proxy-middleware'

const { PORT, NODE_ENV } = process.env
const dev = NODE_ENV === 'development'

polka() // You can also use Express
    .use(
        createProxyMiddleware('/api', {
            changeOrigin: true,
            logLevel: 'debug',
            target: 'http://localhost:9876',
            pathRewrite: {
                '^/api': ''
            }
        }),
        compression({ threshold: 0 }),
        sirv('static', { dev }),
        sapper.middleware(),
    )
    .listen(PORT, (err) => {
        if (err) console.log('error', err)
    })
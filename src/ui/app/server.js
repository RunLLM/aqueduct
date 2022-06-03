require('dotenv').config()

const { createServer } = require('http')
const { parse } = require('url')
const next = require('next')
const nextConfig = require('./next.config.js');

const dev = process.env.NODE_ENV === 'dev'
const hostname = 'localhost'
const port = 3000

const app = next({ dev, hostname, port, dir: __dirname, conf: nextConfig })
const handle = app.getRequestHandler()

app.prepare().then(() => {
  createServer(async (req, res) => {
    try {
      // Be sure to pass `true` as the second argument to `url.parse`.
      // This tells it to parse the query portion of the URL.
      const parsedUrl = parse(req.url, true)
      await handle(req, res, parsedUrl)
    } catch (err) {
      console.error('Error occurred handling', req.url, err)
      res.statusCode = 500
      res.end('Internal server error.')
    }
  }).listen(port, (err) => {
    if (err) throw err
    console.log(`> Ready on http://${hostname}:${port}`)
  })
})

const { createServer } = require('http');
const { parse } = require('url');
const next = require('next');
const nextConfig = require('./next.config');

console.log(__dirname);

const dev = process.env.NODE_ENV === 'dev';
console.log('dev is', dev);
const hostname = 'localhost';
const port = 3000;

const app = next({ dev, hostname, port, dir: __dirname, conf: nextConfig });
const handle = app.getRequestHandler();

process.env.SERVER_ADDRESS = 'localhost:8080';
process.env.NEXT_PUBLIC_PROTOCOL = 'http';

console.log('Process.env', process.env);

app.prepare().then(() => {
    console.log('app.prepare then block');
    process.env.SERVER_ADDRESS = 'localhost:8080';
    process.env.NEXT_PUBLIC_PROTOCOL = 'http';
    console.log('Process.env after prepare', process.env);
    createServer(async (req, res) => {
        try {
            // Be sure to pass `true` as the second argument to `url.parse`.
            // This tells it to parse the query portion of the URL.
            const parsedUrl = parse(req.url, true);
            const { pathname, query } = parsedUrl;
            await handle(req, res, parsedUrl);
        } catch (err) {
            console.error('Error occurred handling', req.url, err);
            res.statusCode = 500;
            res.end('internal server error');
            //await app.render(req, res, '/login', query);
            //await handle(req, res, parsedUrl);
        }
    }).listen(port, (err) => {
        if (err) throw err;
        console.log(`> Ready on http://${hostname}:${port}`);
    });
});

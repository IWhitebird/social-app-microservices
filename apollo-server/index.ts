import express from 'express';
import { createProxyMiddleware } from 'http-proxy-middleware';
import cors from 'cors';

const app = express();

// Enable CORS
app.use(cors());

// Set up proxy to the Go GraphQL server
app.use('/graphql', createProxyMiddleware({
  target: 'http://localhost:8080',
  changeOrigin: true,
}));

// Add a simple home page
app.get('/', (req, res) => {
  res.send(`
    <html>
      <head>
        <title>GraphQL Explorer</title>
        <style>
          body { font-family: Arial, sans-serif; margin: 20px; }
          h1 { color: #333; }
          .link { margin: 10px 0; }
          a { color: #0077cc; text-decoration: none; }
          a:hover { text-decoration: underline; }
        </style>
      </head>
      <body>
        <h1>GraphQL API Explorer</h1>
        <div class="link">
          <a href="/graphql" target="_blank">Open GraphQL Endpoint</a>
        </div>
        <p>Your GraphQL API is running at <code>http://localhost:4000/graphql</code></p>
      </body>
    </html>
  `);
});

// Start the server
const PORT = 4000;
app.listen(PORT, () => {
  console.log(`🚀 Server ready at http://localhost:${PORT}`);
  console.log(`🔗 Connected to GraphQL backend at http://localhost:8080/graphql`);
});
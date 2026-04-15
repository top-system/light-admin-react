/**
 * Dev-server proxy. Only effective in `umi dev`; production requires your
 * ingress / nginx to forward `/api/v1` and `/ws` to the Go backend.
 */
const backend = 'http://localhost:9999';

const sharedProxy = {
  '/api/v1': {
    target: backend,
    changeOrigin: true,
  },
  '/ws': {
    target: backend.replace(/^http/, 'ws'),
    ws: true,
    changeOrigin: true,
  },
};

export default {
  dev: sharedProxy,
  test: sharedProxy,
  pre: sharedProxy,
};

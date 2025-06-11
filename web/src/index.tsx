import { serve } from "bun";
import index from "./index.html";

const API_BASE_URL = process.env.API_BASE_URL || "http://localhost:8080";

// Helper function to proxy requests
async function proxyRequest(req: Request, pathname: string, search: string) {
  const proxyUrl = new URL(pathname + search, API_BASE_URL);

  try {
    const response = await fetch(proxyUrl, {
      method: req.method,
      headers: req.headers,
      body: req.body,
    });

    // Clone response to modify headers if needed
    const headers = new Headers(response.headers);
    headers.delete("content-encoding"); // Remove compression for development

    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers,
    });
  } catch (error) {
    console.error("Proxy error:", error);
    return new Response(JSON.stringify({ error: "Proxy error" }), {
      status: 502,
      headers: { "Content-Type": "application/json" },
    });
  }
}

const server = serve({
  routes: {
    // Proxy API requests to Go backend
    "/api/*": async (req) => {
      const url = new URL(req.url);
      return proxyRequest(req, url.pathname, url.search);
    },

    // WebSocket endpoint
    "/ws": async (req) => {
      const url = new URL(req.url);

      // Handle WebSocket upgrade
      if (req.headers.get("upgrade") === "websocket") {
        // For now, return a 501 as WebSocket proxying needs special handling
        return new Response("WebSocket proxying not implemented", {
          status: 501,
        });
      }

      return proxyRequest(req, url.pathname, url.search);
    },

    // Serve index.html for all other routes
    "/*": index,
  },

  development: process.env.NODE_ENV !== "production",
});

console.log(`🚀 Server running at ${server.url}`);
console.log(`📡 Proxying API requests to ${API_BASE_URL}`);

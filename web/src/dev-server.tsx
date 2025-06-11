#!/usr/bin/env bun
import { transpiler } from "bun";

const API_BASE_URL = process.env.API_BASE_URL || "http://localhost:8080";

Bun.serve({
  port: 3000,
  async fetch(req) {
    const url = new URL(req.url);

    // Proxy API requests
    if (url.pathname.startsWith("/api/") || url.pathname === "/ws") {
      const proxyUrl = new URL(url.pathname + url.search, API_BASE_URL);

      try {
        const response = await fetch(proxyUrl, {
          method: req.method,
          headers: req.headers,
          body: req.body,
        });

        return new Response(response.body, {
          status: response.status,
          headers: response.headers,
        });
      } catch (error) {
        console.error("Proxy error:", error);
        return new Response("Proxy error", { status: 502 });
      }
    }

    // For .tsx/.ts files, transpile them
    if (url.pathname.endsWith(".tsx") || url.pathname.endsWith(".ts")) {
      const filePath = `src${url.pathname}`;
      const file = Bun.file(filePath);

      if (await file.exists()) {
        const code = await file.text();
        const transpiled = await transpiler.transform(code, "tsx");

        return new Response(transpiled, {
          headers: {
            "Content-Type": "application/javascript",
          },
        });
      }
    }

    // For CSS and other static files
    if (url.pathname !== "/" && !url.pathname.includes("..")) {
      const filePath = `src${url.pathname}`;
      const file = Bun.file(filePath);

      if (await file.exists()) {
        return new Response(file);
      }
    }

    // Serve index.html for all other routes
    return new Response(Bun.file("src/index.html"), {
      headers: { "Content-Type": "text/html" },
    });
  },
});

console.log("Dev server running at http://localhost:3000");
console.log("Proxying API requests to", API_BASE_URL);

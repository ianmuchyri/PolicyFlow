import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  trailingSlash: false,
  // Required for static export with the Go binary
  images: {
    unoptimized: true,
  },
};

export default nextConfig;

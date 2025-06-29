/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
  async rewrites() {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    return [
      {
        source: '/api/proxy/:path*',
        destination: `${apiUrl}/:path*`,
      },
    ]
  },
}

export default nextConfig

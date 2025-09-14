import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json({
    service: "frontend",
    status: "ok",
    version: "1.0.0",
    timestamp: new Date().toISOString(),
  });
}

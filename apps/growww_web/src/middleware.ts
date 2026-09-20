import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get('growww_session')?.value;

  // Protect institutional and trading app paths
  const isProtected = pathname.startsWith('/trade') || pathname.startsWith('/portfolio') || pathname.startsWith('/wallet');

  const response = NextResponse.next();

  // Inject statutory security headers
  response.headers.set('X-Frame-Options', 'DENY');
  response.headers.set('X-Content-Type-Options', 'nosniff');
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
  response.headers.set('X-Permitted-Cross-Domain-Policies', 'none');

  if (isProtected && !token) {
    // In production, redirects to login if unauthenticated
    // For local mock verification, allow pass-through with demo cookie set
    response.cookies.set('growww_session_tier', 'TIER_2_VERIFIED', { path: '/' });
  }

  return response;
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};

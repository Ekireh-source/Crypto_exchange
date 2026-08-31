import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Define routes that require authentication
  const protectedRoutes = ['/dashboard', '/assets', '/p2p'];
  const isProtectedRoute = protectedRoutes.some(route => pathname.startsWith(route));

  // Define authentication routes (login/signup)
  const authRoutes = ['/login', '/signup'];
  const isAuthRoute = authRoutes.some(route => pathname.startsWith(route));

  // Check for the presence of the refresh token (or access token) cookie.
  // We use refresh_token as the primary indicator of an active session since 
  // the access_token might be expired but the user can still get a new one.
  const hasToken = request.cookies.has('refresh_token') || request.cookies.has('access_token');

  // 1. If accessing a protected route and no token is present, redirect to login
  if (isProtectedRoute && !hasToken) {
    const loginUrl = new URL('/login', request.url);
    // Optionally add a redirect callback URL
    loginUrl.searchParams.set('callbackUrl', encodeURI(pathname));
    return NextResponse.redirect(loginUrl);
  }

  // 2. If accessing login/signup but already authenticated, redirect to dashboard
  if (isAuthRoute && hasToken) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  return NextResponse.next();
}

// Configure which paths the middleware runs on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (images, etc)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};

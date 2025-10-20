import Link from 'next/link';

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="text-center space-y-6">
        <h1 className="text-4xl font-bold tracking-tight">
          SaaS Admin Dashboard
        </h1>
        <p className="text-xl text-muted-foreground">
          Platform Administration for Beakon Status Page
        </p>

        <div className="flex gap-4 justify-center mt-8">
          <Link
            href="/admin/dashboard"
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg font-semibold hover:opacity-90 transition-opacity"
          >
            Go to Dashboard
          </Link>
          <Link
            href="/login"
            className="px-6 py-3 border border-border rounded-lg font-semibold hover:bg-muted transition-colors"
          >
            Login
          </Link>
        </div>

        <div className="mt-12 p-6 border border-border rounded-lg bg-card max-w-2xl">
          <h2 className="text-lg font-semibold mb-2">
            Next.js Frontend - Setup Complete ✅
          </h2>
          <ul className="text-sm text-left space-y-1 text-muted-foreground">
            <li>✓ Next.js 14 with App Router</li>
            <li>✓ TypeScript configured</li>
            <li>✓ Tailwind CSS ready</li>
            <li>✓ API proxy to Go backend (localhost:8098)</li>
            <li>✓ Ready for shadcn/ui components</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

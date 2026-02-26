import Link from "next/link";
import {
  Shield,
  Lock,
  FileCheck,
  Users,
  GitBranch,
  Mail,
  Database,
  Package,
  ArrowRight,
  CheckCircle,
  Terminal,
  Zap,
  Building2,
} from "lucide-react";

// ─── WIP Banner ────────────────────────────────────────────────────────────

function WIPBanner() {
  return (
    <div className="w-full bg-amber-50 dark:bg-amber-950/40 border-b border-amber-200 dark:border-amber-800/50 py-2.5 px-4">
      <p className="text-center text-sm text-amber-800 dark:text-amber-300 font-medium">
        🚧 <span className="font-semibold">Work in Progress</span> — Core
        features are functional. Production hardening and SSO are on the
        roadmap.
      </p>
    </div>
  );
}

// ─── Hero ──────────────────────────────────────────────────────────────────

function Hero() {
  return (
    <section className="relative overflow-hidden pt-20 pb-24 px-4">
      {/* Grid background */}
      <div
        className="absolute inset-0 opacity-[0.03] dark:opacity-[0.06]"
        style={{
          backgroundImage:
            "repeating-linear-gradient(0deg, transparent, transparent 39px, #888 39px, #888 40px), repeating-linear-gradient(90deg, transparent, transparent 39px, #888 39px, #888 40px)",
        }}
      />

      {/* Glow */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-150 h-75 bg-blue-500/10 dark:bg-blue-500/15 rounded-full blur-3xl pointer-events-none" />

      <div className="relative max-w-4xl mx-auto text-center">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 text-xs font-medium mb-8">
          <Zap className="h-3 w-3" />
          Open source · Self-hosted · No SaaS lock-in
        </div>

        {/* Logo + Title */}
        <div className="flex items-center justify-center gap-4 mb-6">
          <div className="shrink-0 w-14 h-14 rounded-2xl bg-blue-600 flex items-center justify-center shadow-xl shadow-blue-600/30">
            <Shield className="h-8 w-8 text-white" />
          </div>
          <h1 className="text-5xl sm:text-6xl font-extrabold tracking-tight text-fd-foreground">
            PolicyFlow
          </h1>
        </div>

        {/* Tagline */}
        <p className="text-xl sm:text-2xl text-fd-muted-foreground max-w-2xl mx-auto leading-relaxed mb-10">
          Policy management that lives{" "}
          <span className="text-fd-foreground font-semibold">
            on your infrastructure
          </span>
          .
          <br />
          One binary. Zero ops. Full audit trail.
        </p>

        {/* CTA Buttons */}
        <div className="flex flex-wrap items-center justify-center gap-3">
          <Link
            href="/docs/quickstart"
            className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold text-sm transition-colors shadow-lg shadow-blue-600/25"
          >
            Get started
            <ArrowRight className="h-4 w-4" />
          </Link>
          <Link
            href="/docs"
            className="inline-flex items-center gap-2 px-6 py-3 border border-fd-border bg-fd-background hover:bg-fd-muted text-fd-foreground rounded-xl font-semibold text-sm transition-colors"
          >
            Documentation
          </Link>
          <a
            href="https://github.com/yourorg/policyflow"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-6 py-3 border border-fd-border bg-fd-background hover:bg-fd-muted text-fd-muted-foreground rounded-xl font-semibold text-sm transition-colors"
          >
            <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
            GitHub
          </a>
        </div>

        {/* Quick install */}
        <div className="mt-12 inline-flex items-center gap-3 px-5 py-3 bg-fd-card border border-fd-border rounded-xl text-sm font-mono text-fd-muted-foreground">
          <Terminal className="h-4 w-4 text-blue-500 shrink-0" />
          <span>
            <span className="text-slate-400">$</span> make build &amp;&amp;
            ./build/policyflow
          </span>
        </div>
      </div>
    </section>
  );
}

// ─── Features ──────────────────────────────────────────────────────────────

const features = [
  {
    icon: Package,
    title: "Single Binary",
    description:
      "The entire application — Go backend and Next.js frontend — ships as a single ~17MB binary. Copy it to any server and run.",
    color: "text-blue-500",
    bg: "bg-blue-500/10 dark:bg-blue-500/15",
  },
  {
    icon: Database,
    title: "SQLite — Zero Ops",
    description:
      "All data lives in a single file. No Postgres, Redis, or infrastructure to manage. Backups are a simple file copy.",
    color: "text-emerald-500",
    bg: "bg-emerald-500/10 dark:bg-emerald-500/15",
  },
  {
    icon: Mail,
    title: "Magic-Link Auth",
    description:
      "No passwords. Staff log in via a secure link emailed to their work address. Admins control who has access.",
    color: "text-violet-500",
    bg: "bg-violet-500/10 dark:bg-violet-500/15",
  },
  {
    icon: Building2,
    title: "Department-Scoped RBAC",
    description:
      "Three roles — SuperAdmin, DeptAdmin, Staff — with server-enforced scoping. Department admins manage only their own team and policies.",
    color: "text-cyan-500",
    bg: "bg-cyan-500/10 dark:bg-cyan-500/15",
  },
  {
    icon: GitBranch,
    title: "Policy Versioning",
    description:
      "Every policy update creates a new version. Staff are notified and must re-acknowledge. Full history is preserved.",
    color: "text-amber-500",
    bg: "bg-amber-500/10 dark:bg-amber-500/15",
  },
  {
    icon: FileCheck,
    title: "Acknowledgement Audit Trail",
    description:
      "Each acknowledgement is recorded with a SHA-256 signature hash tied to the user, version, and timestamp.",
    color: "text-pink-500",
    bg: "bg-pink-500/10 dark:bg-pink-500/15",
  },
  {
    icon: Lock,
    title: "Self-Hosted & Private",
    description:
      "Your policy content and employee data never leaves your infrastructure. Apache 2.0 licensed — own it completely.",
    color: "text-slate-500",
    bg: "bg-slate-500/10 dark:bg-slate-500/15",
  },
];

function Features() {
  return (
    <section className="py-20 px-4 bg-fd-muted/30">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-3xl font-bold text-fd-foreground mb-3">
            Everything you need, nothing you don&apos;t
          </h2>
          <p className="text-fd-muted-foreground max-w-xl mx-auto">
            PolicyFlow is purpose-built for simplicity. No cloud dependencies,
            no subscriptions, no surprise bills.
          </p>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {features.map((f) => (
            <div
              key={f.title}
              className="bg-fd-card border border-fd-border rounded-2xl p-6 hover:border-blue-300 dark:hover:border-blue-700 transition-colors group"
            >
              <div
                className={`w-10 h-10 rounded-xl ${f.bg} flex items-center justify-center mb-4`}
              >
                <f.icon className={`h-5 w-5 ${f.color}`} />
              </div>
              <h3 className="font-semibold text-fd-foreground mb-2">
                {f.title}
              </h3>
              <p className="text-sm text-fd-muted-foreground leading-relaxed">
                {f.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ─── Workflow ──────────────────────────────────────────────────────────────

const steps = [
  {
    step: "01",
    actor: "SuperAdmin",
    title: "Set up departments & admins",
    description:
      "Create departments and assign Department Admins. Set organization-wide policies that apply to everyone.",
  },
  {
    step: "02",
    actor: "DeptAdmin",
    title: "Manage team & dept policies",
    description:
      "Add team members and publish department-scoped policies. Scoping is enforced server-side.",
  },
  {
    step: "03",
    actor: "Staff",
    title: "Read & acknowledge",
    description:
      "Staff log in via magic link, read their policies, and click Acknowledge. A cryptographic record is created.",
  },
  {
    step: "04",
    actor: "Admin",
    title: "Track compliance",
    description:
      "The dashboard shows who has acknowledged each policy version. See gaps at a glance.",
  },
];

function Workflow() {
  return (
    <section className="py-20 px-4">
      <div className="max-w-5xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-3xl font-bold text-fd-foreground mb-3">
            Simple by design
          </h2>
          <p className="text-fd-muted-foreground">
            Four steps from setup to full compliance
          </p>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {steps.map((s, i) => (
            <div key={s.step} className="relative">
              {/* Connector line */}
              {i < steps.length - 1 && (
                <div className="hidden lg:block absolute top-6 left-full w-6 h-px bg-fd-border" />
              )}
              <div className="bg-fd-card border border-fd-border rounded-2xl p-5 overflow-hidden">
                <div className="flex items-start justify-between mb-3">
                  <span className="text-3xl font-black text-blue-500/20 dark:text-blue-400/25 leading-none">
                    {s.step}
                  </span>
                  <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-fd-muted text-fd-muted-foreground">
                    {s.actor}
                  </span>
                </div>
                <h3 className="font-semibold text-fd-foreground text-sm mb-1.5">
                  {s.title}
                </h3>
                <p className="text-xs text-fd-muted-foreground leading-relaxed">
                  {s.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ─── Tech Stack / Architecture snippet ─────────────────────────────────────

function Architecture() {
  return (
    <section className="py-20 px-4 bg-fd-muted/30">
      <div className="max-w-5xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
          {/* Text */}
          <div>
            <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 text-xs font-mono mb-6">
              tech stack
            </div>
            <h2 className="text-3xl font-bold text-fd-foreground mb-4">
              Built on boring technology
            </h2>
            <p className="text-fd-muted-foreground mb-6 leading-relaxed">
              PolicyFlow is built on proven, boring technology — the kind that
              runs reliably for years without intervention.
            </p>
            <ul className="space-y-3">
              {[
                ["Go 1.25", "Fast, statically compiled backend"],
                ["Echo v4", "HTTP framework with clean middleware"],
                ["SQLite (WAL)", "Single-file database, zero setup"],
                [
                  "Next.js 15",
                  "Static export — no server-side rendering needed",
                ],
                ["Tailwind CSS v4", "Utility-first styling"],
                ["JWT (HMAC-SHA256)", "Stateless session tokens"],
              ].map(([name, desc]) => (
                <li key={name} className="flex items-start gap-3">
                  <CheckCircle className="h-4 w-4 text-emerald-500 mt-0.5 shrink-0" />
                  <span className="text-sm text-fd-foreground">
                    <span className="font-mono font-medium">{name}</span>
                    <span className="text-fd-muted-foreground"> — {desc}</span>
                  </span>
                </li>
              ))}
            </ul>
          </div>

          {/* Code block */}
          <div className="rounded-2xl bg-slate-900 dark:bg-slate-950 border border-slate-700 overflow-hidden shadow-2xl">
            <div className="flex items-center gap-1.5 px-4 py-3 border-b border-slate-700/80 bg-slate-800/50">
              <div className="w-2.5 h-2.5 rounded-full bg-red-500/70" />
              <div className="w-2.5 h-2.5 rounded-full bg-amber-500/70" />
              <div className="w-2.5 h-2.5 rounded-full bg-green-500/70" />
              <span className="ml-2 text-xs text-slate-400 font-mono">
                main.go
              </span>
            </div>
            <pre className="text-xs font-mono p-5 text-slate-300 leading-relaxed overflow-x-auto">
              <code>{`//go:embed all:web/out
var webFiles embed.FS

func main() {
  db := database.New(sqlDB)
  db.Init()
  seed.Run(db)

  e := echo.New()
  // JWT-protected API routes
  api.POST("/magic-link", authH.RequestMagicLink)
  api.GET("/magic-login",  authH.MagicLogin)
  auth.GET("/policies",    policyH.List)
  auth.POST("/:id/acknowledge", policyH.Acknowledge)

  // Embedded SPA fallback
  e.GET("/*", spaHandler(webFiles))

  e.Start(":8080")
}`}</code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}

// ─── Roadmap ───────────────────────────────────────────────────────────────

const roadmap = [
  { label: "SSO / OIDC", done: false },
  { label: "LDAP / Active Directory sync", done: false },
  { label: "Employee CSV import", done: false },
  { label: "PDF audit log export", done: false },
  { label: "Rich text editor", done: false },
  { label: "Email reminders for unacknowledged policies", done: false },
  { label: "Webhook notifications (Slack, Teams)", done: false },
  { label: "Refresh token rotation", done: false },
];

function Roadmap() {
  return (
    <section className="py-20 px-4">
      <div className="max-w-3xl mx-auto text-center">
        <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-xs font-medium mb-6">
          🚧 Roadmap
        </div>
        <h2 className="text-3xl font-bold text-fd-foreground mb-3">
          What&apos;s coming
        </h2>
        <p className="text-fd-muted-foreground mb-10">
          PolicyFlow is an MVP. Here&apos;s what&apos;s planned for future
          releases.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-left">
          {roadmap.map((item) => (
            <div
              key={item.label}
              className="flex items-center gap-3 px-4 py-3 rounded-xl border border-fd-border bg-fd-card text-sm"
            >
              <div className="w-4 h-4 rounded border-2 border-fd-muted-foreground/40 shrink-0" />
              <span className="text-fd-muted-foreground">{item.label}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ─── CTA ───────────────────────────────────────────────────────────────────

function CTA() {
  return (
    <section className="py-20 px-4">
      <div className="max-w-2xl mx-auto text-center">
        <div className="relative rounded-3xl bg-linear-to-br from-blue-600 to-blue-700 p-12 overflow-hidden shadow-2xl shadow-blue-600/20">
          {/* Background decoration */}
          <div className="absolute top-0 right-0 w-48 h-48 bg-white/5 rounded-full -translate-y-1/2 translate-x-1/2" />
          <div className="absolute bottom-0 left-0 w-32 h-32 bg-white/5 rounded-full translate-y-1/2 -translate-x-1/2" />

          <Shield className="h-10 w-10 text-white/80 mx-auto mb-6" />
          <h2 className="text-3xl font-bold text-white mb-4">
            Own your compliance data
          </h2>
          <p className="text-blue-100 mb-8 leading-relaxed">
            Deploy PolicyFlow on your own servers in minutes. No vendor lock-in,
            no monthly bill, no data leaving your network.
          </p>
          <div className="flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/docs/quickstart"
              className="inline-flex items-center gap-2 px-6 py-3 bg-white text-blue-700 hover:bg-blue-50 rounded-xl font-semibold text-sm transition-colors shadow-lg"
            >
              Read the quickstart
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link
              href="/docs"
              className="inline-flex items-center gap-2 px-6 py-3 bg-blue-500/40 hover:bg-blue-500/60 border border-white/20 text-white rounded-xl font-semibold text-sm transition-colors"
            >
              Browse the docs
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

// ─── Footer ────────────────────────────────────────────────────────────────

function Footer() {
  return (
    <footer className="border-t border-fd-border py-8 px-4">
      <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-fd-muted-foreground">
        <div className="flex items-center gap-2">
          <Shield className="h-4 w-4 text-blue-500" />
          <span className="font-medium text-fd-foreground">PolicyFlow</span>
          <span>— Apache 2.0</span>
        </div>
        <div className="flex items-center gap-6">
          <Link
            href="/docs"
            className="hover:text-fd-foreground transition-colors"
          >
            Docs
          </Link>
          <Link
            href="/docs/api"
            className="hover:text-fd-foreground transition-colors"
          >
            API
          </Link>
          <Link
            href="/docs/deployment"
            className="hover:text-fd-foreground transition-colors"
          >
            Deploy
          </Link>
          <a
            href="https://github.com/yourorg/policyflow"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-fd-foreground transition-colors"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-screen">
      <WIPBanner />
      <Hero />
      <Features />
      <Workflow />
      <Architecture />
      <Roadmap />
      <CTA />
      <Footer />
    </div>
  );
}

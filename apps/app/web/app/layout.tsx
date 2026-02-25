import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "PolicyFlow",
  description: "Self-hosted policy management",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-slate-50 dark:bg-slate-900">
        {children}
      </body>
    </html>
  );
}

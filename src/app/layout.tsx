import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Nav } from "@/components/nav";
import { AuthProvider } from "@/components/auth-provider";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "ThumbTrend - YouTube Thumbnail Trend Analyzer",
  description:
    "Discover trending YouTube thumbnail patterns, micro-genres, and visual insights powered by AI.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} dark h-full antialiased`}
    >
      <body className="flex min-h-full flex-col bg-background text-foreground">
        <AuthProvider>
          <Nav />
          <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-6">
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  );
}

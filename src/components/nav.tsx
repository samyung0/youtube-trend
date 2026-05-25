"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuth } from "@/components/auth-provider";
import { BarChart3, Video, Hash, Users, PenTool, LogIn, LogOut } from "lucide-react";

const exploreTabs = [
  { href: "/", label: "Trending Videos", icon: Video },
  { href: "/topics", label: "Trending Topics", icon: Hash },
  { href: "/channels", label: "Trending Channels", icon: Users },
];

export function Nav() {
  const pathname = usePathname();
  const { user, login, logout } = useAuth();

  const isExplore = !pathname.startsWith("/create");

  return (
    <header className="sticky top-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-red-600">
            <BarChart3 className="h-3.5 w-3.5 text-white" />
          </div>
          <span className="text-lg font-bold tracking-tight">
            Thumb<span className="text-red-500">Trend</span>
          </span>
        </Link>

        <div className="flex items-center gap-1">
          <Link
            href="/"
            className={cn(
              "rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
              isExplore
                ? "bg-accent text-accent-foreground"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            Explore Trends
          </Link>
          <Link
            href="/create"
            className={cn(
              "flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors",
              pathname.startsWith("/create")
                ? "bg-accent text-accent-foreground"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <PenTool className="h-3.5 w-3.5" />
            Create
          </Link>
        </div>

        <div className="flex items-center gap-2">
          {user ? (
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">{user.name}</span>
              <button
                onClick={logout}
                className="flex items-center gap-1 rounded-lg px-2 py-1.5 text-sm text-muted-foreground hover:text-foreground"
              >
                <LogOut className="h-3.5 w-3.5" />
              </button>
            </div>
          ) : (
            <button
              onClick={login}
              className="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm font-medium hover:bg-accent/80"
            >
              <LogIn className="h-3.5 w-3.5" />
              Sign in
            </button>
          )}
        </div>
      </div>

      {isExplore && (
        <div className="border-t border-border/30 bg-background/50">
          <div className="mx-auto flex max-w-7xl items-center gap-1 px-4 py-1">
            {exploreTabs.map(({ href, label, icon: Icon }) => (
              <Link
                key={href}
                href={href}
                className={cn(
                  "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors",
                  pathname === href
                    ? "bg-accent font-medium text-accent-foreground"
                    : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                )}
              >
                <Icon className="h-3.5 w-3.5" />
                {label}
              </Link>
            ))}
          </div>
        </div>
      )}
    </header>
  );
}

"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { setToken } from "@/lib/api";

function CallbackHandler() {
  const router = useRouter();
  const params = useSearchParams();

  useEffect(() => {
    const token = params.get("token");
    if (token) {
      setToken(token);
      router.replace("/");
    } else {
      router.replace("/?error=auth_failed");
    }
  }, [params, router]);

  return (
    <div className="flex items-center justify-center py-20">
      <p className="text-muted-foreground">Signing in...</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center py-20">
          <p className="text-muted-foreground">Signing in...</p>
        </div>
      }
    >
      <CallbackHandler />
    </Suspense>
  );
}

"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import AuthCard from "../_components/AuthCard";
import { Icon } from "@iconify/react";
import Link from "next/link";
import { authService } from "@/feature/auth/auth.service";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [message, setMessage] = useState("Verifying your email...");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setMessage("Verification token is missing.");
      return;
    }

    const verifyToken = async () => {
      try {
        await authService.verifyEmail(token);
        setStatus("success");
        setMessage("Your email has been successfully verified!");
      } catch (error: any) {
        setStatus("error");
        setMessage(error.response?.data?.error || error.response?.data?.message || "Failed to verify email. The link may have expired.");
      }
    };

    verifyToken();
  }, [token]);

  return (
    <div className="flex flex-col items-center justify-center py-8 space-y-6 text-center">
      {status === "loading" && (
        <Icon icon="hugeicons:loading-02" className="size-16 text-indigo-400 animate-spin" />
      )}
      
      {status === "success" && (
        <Icon icon="hugeicons:tick-circle" className="size-16 text-green-500" />
      )}
      
      {status === "error" && (
        <Icon icon="hugeicons:cancel-circle" className="size-16 text-red-500" />
      )}

      <p className="text-slate-400 text-[0.95rem]">{message}</p>

      {status !== "loading" && (
        <Link 
          href="/login"
          className="w-full h-[52px] mt-4 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white text-[0.95rem] font-medium tracking-wide flex items-center justify-center gap-2 shadow-[0_4px_24px_rgba(79,70,229,0.15)] transition-all duration-200 active:scale-[0.99]"
        >
          Go to Login
          <Icon icon="lucide:arrow-right" className="text-lg ml-1" />
        </Link>
      )}
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <AuthCard
      title="Email Verification"
      subtitle="Verification status"
    >
      <Suspense fallback={
        <div className="flex flex-col items-center justify-center py-8 space-y-6 text-center">
          <Icon icon="lucide:loader-2" className="size-12 text-indigo-400 animate-spin" />
          <p className="text-slate-400 text-[0.95rem]">Loading...</p>
        </div>
      }>
        <VerifyEmailContent />
      </Suspense>
    </AuthCard>
  );
}

"use client";

import FileUpload from "@/components/file-upload/file-upload";
import { createClient } from "@/utils/supabase/client";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { FolderOpen } from "lucide-react";

export default function DashboardPage() {
  const router = useRouter();
  const supabase = createClient();

  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    async function isUserAuthenticated() {
      const { data, error } = await supabase.auth.getUser();
      if (error || !data.user) {
        router.push("/sign-in");
      } else {
        setUser(data.user);
      }
    }
    isUserAuthenticated();
  }, [supabase, router]);

  if (!user) return <p>Loading...</p>;

  return (
    <div className="flex flex-col items-center p-6">
      <div className="w-full max-w-lg flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Button
          variant="outline"
          onClick={() => router.push("/files")}
          className="flex items-center gap-2"
        >
          <FolderOpen className="w-4 h-4" />
          View Files
        </Button>
      </div>
      <FileUpload />
    </div>
  );
}

"use client";

import FileList from "@/components/file-list/file-list";
import { createClient } from "@/utils/supabase/client";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

export default function FilesPage() {
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
    <div className="flex flex-col items-center">
      <FileList />
    </div>
  );
} 
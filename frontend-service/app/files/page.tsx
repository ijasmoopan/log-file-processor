"use client";

import FileList from "@/components/file-list/file-list";
import { createClient } from "@/utils/supabase/client";
import { User } from "@supabase/supabase-js";
import { useRouter } from "next/navigation";
import { useEffect, useState, useCallback } from "react";

interface FileProgress {
  fileName: string;
  clientId: string;
  progress: number;
  status: string;
  processedAt: string;
  error?: string;
}

interface FileProgressResponse {
  file_name: string;
  client_id: string;
  progress: number;
  status: string;
  processed_at: string;
  error?: string;
}

export default function FilesPage() {
  const router = useRouter();
  const supabase = createClient();
  const [user, setUser] = useState<User | null>(null);
  const [fileProgress, setFileProgress] = useState<
    Record<string, FileProgress>
  >({});
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);

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

  const connectWebSocket = () => {
    if (!user?.id) return; // Don't connect if no user ID
    if (ws?.readyState === WebSocket.OPEN) return; // Already connected

    const socket = new WebSocket(
      `ws://localhost:8080/api/v1/ws?client_id=${user.id}`
    );

    socket.onopen = () => {
      console.log("WebSocket connection established");
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      try {
        let data: FileProgressResponse;
        try {
          data = JSON.parse(event.data);
        } catch (error) {
          const jsonObjects = event.data.match(/\{[^{}]+\}/g) || [];
          const parsedObjects: FileProgressResponse[] = jsonObjects.map(
            (json: any) => JSON.parse(json)
          );
          data = parsedObjects[0];
        }
        setFileProgress((prev) => ({
          ...prev,
          [data.file_name]: {
            fileName: data.file_name,
            clientId: data.client_id,
            progress: data.progress,
            status: data.status,
            processedAt: data.processed_at,
            error: data?.error,
          },
        }));

        // If file is completed or has error, check if all files are done
        if (data.status === "completed" || data.status === "error") {
          setTimeout(() => {
            const hasActiveFiles = Object.values(fileProgress).some(
              (file) => file.status === "processing"
            );
            if (!hasActiveFiles) {
              disconnectWebSocket();
            }
          }, 1000); // Give a small delay to ensure all messages are received
        }
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    socket.onclose = () => {
      console.log("WebSocket connection closed");
      setWs(null);
      setIsConnected(false);
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      setIsConnected(false);
    };

    setWs(socket);
  };

  useEffect(() => {
    if (user?.id) {
      connectWebSocket();
    }
  }, [user?.id]);

  const disconnectWebSocket = useCallback(() => {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.close();
      setWs(null);
      setFileProgress({});
      setIsConnected(false);
    }
  }, [ws]);

  if (!user) return <p>Loading...</p>;

  return (
    <div className="flex flex-col items-center">
      <FileList
        user={user}
        fileProgress={fileProgress}
        onProcessingComplete={disconnectWebSocket}
        isConnected={isConnected}
      />
    </div>
  );
}

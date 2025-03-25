"use client";

import React, { useEffect, useState } from "react";
import { format } from "date-fns";
import { FileText, Calendar, Clock, AlertCircle, CheckCircle, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface FileResult {
  ID: number;
  FileName: string;
  ClientID: string;
  Status: string;
  WarnCount?: number;
  ErrorCount?: number;
  Error?: string;
  CreatedAt: string;
  UpdatedAt: string;
}

interface FileDetailsProps {
  file: {
    file_name: string;
    size: number;
    path: string;
    upload_time: string;
    last_modified: string;
  };
  progress?: {
    fileName: string;
    clientId: string;
    progress: number;
    status: string;
    processedAt: string;
    error?: string;
  };
  isOpen: boolean;
  onClose: () => void;
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

export default function FileDetails({
  file,
  progress,
  isOpen,
  onClose,
}: FileDetailsProps) {
  const [fileResult, setFileResult] = useState<FileResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFileResult = async () => {
      if (!isOpen) return;
      
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(`http://localhost:8080/api/v1/results/filename/${encodeURIComponent(file.file_name)}`);
        
        if (!response.ok) {
          if (response.status === 404) {
            // No results found is not an error
            setFileResult(null);
            return;
          }
          throw new Error("Failed to fetch file results");
        }

        const data = await response.json();
        setFileResult(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred");
        console.error("Error fetching file results:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchFileResult();
  }, [file.file_name, isOpen]);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileText className="w-5 h-5" />
            File Details
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="space-y-2">
            <h3 className="font-medium text-lg">{file.file_name}</h3>
            <div className="grid gap-2 text-sm text-gray-500">
              <div className="flex items-center gap-2">
                <span className="font-medium">Size:</span>
                {formatFileSize(file.size)}
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4" />
                <span>Uploaded:</span>
                {format(new Date(file.upload_time), "PPp")}
              </div>
              <div className="flex items-center gap-2">
                <Clock className="w-4 h-4" />
                <span>Modified:</span>
                {format(new Date(file.last_modified), "PPp")}
              </div>
            </div>
          </div>
          
          <div className="space-y-2 border-t pt-4">
            <h4 className="font-medium">Processing Results</h4>
            {loading ? (
              <div className="flex items-center gap-2 text-sm text-gray-500">
                <Loader2 className="w-4 h-4 animate-spin" />
                Loading results...
              </div>
            ) : error ? (
              <div className="flex items-center gap-2 text-sm text-red-500">
                <AlertCircle className="w-4 h-4" />
                {error}
              </div>
            ) : fileResult ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  {fileResult.Status === "completed" ? (
                    <CheckCircle className="w-5 h-5 text-green-500" />
                  ) : fileResult.Status === "error" ? (
                    <AlertCircle className="w-5 h-5 text-red-500" />
                  ) : null}
                  <span className="font-medium capitalize">{fileResult.Status}</span>
                </div>
                {fileResult.WarnCount !== undefined && (
                  <p className="text-sm text-yellow-600">
                    Warnings: {fileResult.WarnCount}
                  </p>
                )}
                {fileResult.ErrorCount !== undefined && (
                  <p className="text-sm text-red-600">
                    Errors: {fileResult.ErrorCount}
                  </p>
                )}
                {fileResult.Error && (
                  <p className="text-sm text-red-500">{fileResult.Error}</p>
                )}
              </div>
            ) : progress ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  {progress.status === "completed" ? (
                    <CheckCircle className="w-5 h-5 text-green-500" />
                  ) : progress.status === "error" ? (
                    <AlertCircle className="w-5 h-5 text-red-500" />
                  ) : null}
                  <span className="font-medium">{progress.status}</span>
                </div>
                {progress.processedAt && (
                  <p className="text-sm text-gray-500">
                    Processed at: {format(new Date(progress.processedAt), "PPp")}
                  </p>
                )}
                {progress.error && (
                  <p className="text-sm text-red-500">{progress.error}</p>
                )}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No processing results found</p>
            )}
          </div>

        </div>
      </DialogContent>
    </Dialog>
  );
} 
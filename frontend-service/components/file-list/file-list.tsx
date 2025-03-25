"use client";

import React, { useEffect, useState } from "react";
import { Card, CardContent } from "../ui/card";
import { Button } from "../ui/button";
import { useToast } from "@/components/ui/use-toast";
import { format } from "date-fns";
import { FileText, Calendar, Clock } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Progress } from "@/components/ui/progress";
import { User } from "@supabase/supabase-js";
import FileDetails from "../file-details/file-details";
import { env } from "@/utils/env";

interface File {
  file_name: string;
  size: number;
  path: string;
  upload_time: string;
  last_modified: string;
}

interface FileProgress {
  fileName: string;
  clientId: string;
  progress: number;
  status: string;
  processedAt: string;
  error?: string;
}

interface FileListProps {
  user: User;
  fileProgress: Record<string, FileProgress>;
  onProcessingComplete: () => void;
  isConnected: boolean;
}

interface FileListResponse {
  files: File[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

interface ProcessResponse {
  message: string;
  results: {
    file_name: string;
    status: string;
    processed_at: string;
    error?: string;
  }[];
}

const ConnectionIndicator = ({ isConnected }: { isConnected: boolean }) => (
  <div className="flex items-center gap-2">
    <div
      className={`w-2.5 h-2.5 rounded-full ${
        isConnected ? "bg-green-500" : "bg-red-500"
      }`}
    />
    <span className="text-xs text-gray-500">
      {isConnected ? "Connected" : "Not Connected"}
    </span>
  </div>
);

const PAGE_SIZE = 5;

export default function FileList({
  user,
  fileProgress,
  onProcessingComplete,
  isConnected,
}: FileListProps) {
  const [files, setFiles] = useState<File[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [hasNext, setHasNext] = useState(false);
  const [hasPrevious, setHasPrevious] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());
  const [processing, setProcessing] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const { toast } = useToast();

  const fetchFiles = async (page: number) => {
    try {
      setLoading(true);
      const response = await fetch(
        `${env.backendUrl}/api/v1/files?page=${page}&page_size=${PAGE_SIZE}`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error("Failed to fetch files");
      }

      const data: FileListResponse = await response.json();
      setFiles(data.files);
      setTotalItems(data.total_items);
      setTotalPages(data.total_pages);
      setHasNext(data.has_next);
      setHasPrevious(data.has_prev);
    } catch (error) {
      console.error("Error fetching files:", error);
      toast({
        title: "Error",
        description: "Failed to fetch files. Please try again later.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFiles(currentPage);
  }, [currentPage]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
    }
  };

  const handleSelectAll = () => {
    if (selectedFiles.size === files.length) {
      setSelectedFiles(new Set());
    } else {
      setSelectedFiles(new Set(files.map((file) => file.file_name)));
    }
  };

  const handleSelectFile = (fileName: string) => {
    const newSelected = new Set(selectedFiles);
    if (newSelected.has(fileName)) {
      newSelected.delete(fileName);
    } else {
      newSelected.add(fileName);
    }
    setSelectedFiles(newSelected);
  };

  const handleProcessFiles = async () => {
    if (selectedFiles.size === 0) {
      toast({
        title: "No files selected",
        description: "Please select at least one file to process.",
        variant: "destructive",
      });
      return;
    }

    try {
      setProcessing(true);

      const response = await fetch(`${env.backendUrl}/api/v1/process`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
        body: JSON.stringify({
          file_names: Array.from(selectedFiles),
          client_id: user?.id || "",
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to process files");
      }

      const data: ProcessResponse = await response.json();

      // Show success message
      toast({
        title: "Success",
        description: `Processing started for ${selectedFiles.size} ${selectedFiles.size === 1 ? "file" : "files"}`,
      });

      // Clear selection after successful processing
      setSelectedFiles(new Set());
    } catch (error) {
      console.error("Error processing files:", error);
      toast({
        title: "Error",
        description: "Failed to process files. Please try again later.",
        variant: "destructive",
      });
      // Disconnect WebSocket on error
      onProcessingComplete();
    } finally {
      setProcessing(false);
    }
  };

  const getFileProgress = (fileName: string) => {
    return fileProgress[fileName];
  };

  const handleFileClick = (file: File) => {
    setSelectedFile(file);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <p>Loading files...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-6">
      <Card className="w-full max-w-4xl mx-auto">
        <CardContent>
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-bold text-black">Uploaded Files</h2>
            <div className="flex items-center gap-4">
              <ConnectionIndicator isConnected={isConnected} />
              {selectedFiles.size > 0 && (
                <Button
                  onClick={handleProcessFiles}
                  disabled={processing}
                  className="bg-blue-600 hover:bg-blue-700 text-white"
                >
                  {processing
                    ? "Processing..."
                    : `Process ${selectedFiles.size} ${selectedFiles.size === 1 ? "File" : "Files"}`}
                </Button>
              )}
            </div>
          </div>

          {files.length === 0 ? (
            <p className="text-center text-gray-500">No files found</p>
          ) : (
            <>
              <div className="space-y-4">
                <div className="flex items-center gap-2 p-2 border-b">
                  <Checkbox
                    id="select-all"
                    checked={selectedFiles.size === files.length}
                    onCheckedChange={handleSelectAll}
                    className="border-gray-300 data-[state=checked]:bg-gray-300 data-[state=checked]:border-gray-300"
                  />
                  <label
                    htmlFor="select-all"
                    className="text-sm font-medium text-gray-700 cursor-pointer"
                  >
                    Select All
                  </label>
                </div>
                {files.map((file) => {
                  const progress = getFileProgress(file.file_name);
                  return (
                    <div
                      key={file.path}
                      className="flex items-start gap-4 p-4 border rounded-lg hover:bg-gray-50 transition-colors cursor-pointer"
                      onClick={() => handleFileClick(file)}
                    >
                      <Checkbox
                        id={file.path}
                        checked={selectedFiles.has(file.file_name)}
                        onCheckedChange={() => handleSelectFile(file.file_name)}
                        onClick={(e) => e.stopPropagation()}
                        className="mt-1 border-gray-300 data-[state=checked]:bg-gray-300 data-[state=checked]:border-gray-300"
                      />
                      <div className="p-2 bg-gray-100 rounded-lg">
                        <FileText className="w-6 h-6 text-gray-600" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <h3 className="font-medium text-lg truncate mb-2 text-black">
                          {file.file_name}
                        </h3>
                        <div className="flex flex-wrap gap-4 text-sm text-gray-500">
                          <div className="flex items-center gap-1">
                            <span className="font-medium">Size:</span>
                            {formatFileSize(file.size)}
                          </div>
                          <div className="flex items-center gap-1">
                            <Calendar className="w-4 h-4" />
                            <span>Uploaded:</span>
                            {format(new Date(file.upload_time), "PP")}
                          </div>
                          <div className="flex items-center gap-1">
                            <Clock className="w-4 h-4" />
                            <span>Modified:</span>
                            {format(new Date(file.last_modified), "PP")}
                          </div>
                        </div>
                        {progress && (
                          <div className="mt-3">
                            <div className="flex justify-between mb-1">
                              <span className="text-sm font-medium text-gray-700">
                                {progress.status}
                              </span>
                              <span className="text-sm font-medium text-gray-700">
                                {progress.progress}%
                              </span>
                            </div>
                            <Progress
                              value={progress.progress}
                              className="h-2"
                            />
                            {progress.processedAt && (
                              <p className="text-sm text-gray-500 mt-1">
                                {progress.processedAt}
                              </p>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>

              <div className="flex justify-center items-center space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={!hasPrevious}
                >
                  Previous
                </Button>
                <span className="px-4">
                  Page {currentPage} of {totalPages} ({totalItems} total files)
                </span>
                <Button
                  variant="outline"
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={!hasNext}
                >
                  Next
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {selectedFile && (
        <FileDetails
          file={selectedFile}
          progress={getFileProgress(selectedFile.file_name)}
          isOpen={!!selectedFile}
          onClose={() => setSelectedFile(null)}
        />
      )}
    </div>
  );
}

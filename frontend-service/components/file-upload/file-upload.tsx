"use client";

import React, { useRef, useState } from "react";
import { Card, CardContent } from "../ui/card";
import { CircleX, UploadCloud } from "lucide-react";
import { Button } from "../ui/button";
import { useToast } from "@/components/ui/use-toast";
import { env } from "@/utils/env";

interface FileWithPreview {
  file: File;
  previewUrl: string | null;
}

interface UploadError {
  message: string;
  details?: string;
}

export default function FileUpload() {
  const [files, setFiles] = useState<FileWithPreview[]>([]);
  const [uploading, setUploading] = useState<boolean>(false);
  const { toast } = useToast();

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = event.target.files;
    if (selectedFiles) {
      const newFiles: FileWithPreview[] = Array.from(selectedFiles).map(file => ({
        file,
        previewUrl: file.type.startsWith("image/") ? URL.createObjectURL(file) : null
      }));
      setFiles(prev => [...prev, ...newFiles]);
    }
  };

  const handleFileRemove = (index: number) => {
    setFiles(prev => {
      const newFiles = [...prev];
      if (newFiles[index].previewUrl) {
        URL.revokeObjectURL(newFiles[index].previewUrl!);
      }
      newFiles.splice(index, 1);
      return newFiles;
    });
  };

  const handleFileUpload = async () => {
    if (files.length === 0) {
      toast({
        title: "No files selected",
        description: "Please select files to upload",
        variant: "destructive",
      });
      return;
    }

    setUploading(true);

    try {
      const formData = new FormData();
      files.forEach(({ file }) => {
        formData.append("files", file);
      });

      const response = await fetch(`${env.backendUrl}/api/v1/upload`, {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        const error: UploadError = {
          message: data.error || "Upload failed",
          details: data.details,
        };
        throw error;
      }

      console.log("Upload response:", data);

      // Clear files after successful upload
      files.forEach(({ previewUrl }) => {
        if (previewUrl) {
          URL.revokeObjectURL(previewUrl);
        }
      });
      setFiles([]);

      toast({
        title: "Success",
        description: "Files uploaded successfully!",
      });
    } catch (error) {
      console.error("Upload error:", error);
      const uploadError = error as UploadError;
      toast({
        title: "Upload Failed",
        description: uploadError.message,
        variant: "destructive",
      });
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center p-6">
      <Card className="w-full max-w-lg p-6 shadow-md">
        <CardContent>
          <div
            className="border-2 border-dashed border-gray-300 p-6 rounded-lg flex flex-col items-center cursor-pointer hover:bg-gray-50"
            onClick={() => fileInputRef.current?.click()}
          >
            <UploadCloud size={40} className="text-gray-500 mb-2" />
            <p className="text-gray-700">Click to upload or drag & drop</p>
            <Button className="mt-3">Select Files</Button>
          </div>

          <input
            title="file-input"
            placeholder="file-input"
            type="file"
            ref={fileInputRef}
            className="hidden"
            onChange={handleFileChange}
            accept=".txt,.log,text/plain,application/x-log"
            multiple
          />

          {files.length > 0 && (
            <div className="mt-4 space-y-4">
              {files.map((fileWithPreview, index) => (
                <div key={index} className="relative">
                  <button
                    title="remove-file"
                    className="absolute -top-3 -right-3 bg-white text-gray-500 rounded-full hover:text-red-600"
                    onClick={() => handleFileRemove(index)}
                  >
                    <CircleX className="w-6 h-6" />
                  </button>

                  {fileWithPreview.previewUrl ? (
                    <img
                      src={fileWithPreview.previewUrl}
                      alt="Preview"
                      className="w-32 h-32 object-cover rounded-md border border-gray-300"
                    />
                  ) : (
                    <div className="p-4 bg-gray-300 rounded-md">
                      {fileWithPreview.file.name}
                    </div>
                  )}
                </div>
              ))}

              <div className="flex justify-center mt-4">
                <Button
                  variant={"outline"}
                  onClick={handleFileUpload}
                  disabled={uploading}
                >
                  {uploading ? "Uploading..." : "Upload Files"}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

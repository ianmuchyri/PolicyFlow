"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface Props {
  content: string;
}

export default function MarkdownRenderer({ content }: Props) {
  return (
    <div className="prose prose-slate dark:prose-invert max-w-none prose-headings:font-semibold prose-a:text-blue-600 dark:prose-a:text-blue-400 prose-pre:bg-slate-800 prose-pre:text-slate-100 prose-code:text-blue-700 dark:prose-code:text-blue-300 prose-code:before:content-none prose-code:after:content-none">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  );
}

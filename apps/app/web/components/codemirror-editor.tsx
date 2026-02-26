"use client";

import CodeMirror from "@uiw/react-codemirror";
import { markdown, markdownLanguage } from "@codemirror/lang-markdown";
import { languages } from "@codemirror/language-data";

interface Props {
  value: string;
  onChange: (value: string) => void;
  height?: string;
}

export default function MarkdownEditor({ value, onChange, height = "320px" }: Props) {
  return (
    <CodeMirror
      value={value}
      height={height}
      extensions={[markdown({ base: markdownLanguage, codeLanguages: languages })]}
      onChange={onChange}
      theme="dark"
      className="rounded-lg overflow-hidden border border-slate-200 dark:border-slate-700 text-sm"
      basicSetup={{
        lineNumbers: true,
        foldGutter: false,
        highlightActiveLine: true,
      }}
    />
  );
}

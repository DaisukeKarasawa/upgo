import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface MarkdownProps {
  children: string;
  className?: string;
}

export default function Markdown({ children, className = '' }: MarkdownProps) {
  return (
    <div className={`prose prose-sm max-w-none ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        skipHtml={true}
        components={{
          a: ({ ...props }) => {
            const href = props.href || '';
            const isSafe = /^(https?|mailto):/.test(href) || href.startsWith('/') || href.startsWith('#');
            if (!isSafe) {
              return <span className="text-gray-400 line-through">{props.children}</span>;
            }
            return (
              <a
                {...props}
                href={href}
                target="_blank"
                rel="noopener noreferrer nofollow"
                aria-label={`${props.children} (opens in new tab)`}
                className="text-gray-700 hover:text-gray-900 transition-colors duration-300"
              />
            );
          },
          code: ({ className, ...props }) => {
            const isInline = !className;
            return isInline ? (
              <code
                {...props}
                className="bg-gray-100 px-1 py-0.5 rounded text-sm font-mono text-gray-800"
              />
            ) : (
              <code {...props} className={className} />
            );
          },
          pre: ({ ...props }) => (
            <pre
              {...props}
              className="bg-gray-50 p-4 rounded overflow-x-auto text-sm"
            />
          ),
          blockquote: ({ ...props }) => (
            <blockquote
              {...props}
              className="border-l-4 border-gray-300 pl-4 italic text-gray-600"
            />
          ),
          table: ({ ...props }) => (
            <div className="overflow-x-auto">
              <table {...props} className="min-w-full border-collapse" />
            </div>
          ),
          th: ({ ...props }) => (
            <th
              {...props}
              className="border border-gray-300 px-4 py-2 bg-gray-100 text-left font-semibold"
            />
          ),
          td: ({ ...props }) => (
            <td
              {...props}
              className="border border-gray-300 px-4 py-2"
            />
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}

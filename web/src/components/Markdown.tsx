import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';

interface MarkdownProps {
  children: string;
  className?: string;
}

// サニタイズスキーマ: 許可するタグと属性をホワイトリスト形式で定義
const sanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    ...(defaultSchema.tagNames || []),
    'details',
    'summary',
    'kbd',
    'sup',
    'sub',
  ],
  attributes: {
    ...defaultSchema.attributes,
    a: [
      ...(defaultSchema.attributes?.a || []),
      ['target', '_blank'],
      ['rel', 'noopener noreferrer nofollow'],
    ],
    img: [
      ['src'],
      ['alt'],
      ['title'],
      ['width'],
      ['height'],
    ],
    code: [
      ['className'],
    ],
    pre: [
      ['className'],
    ],
  },
  protocols: {
    ...defaultSchema.protocols,
    href: ['http', 'https', 'mailto'],
    src: ['http', 'https'],
  },
};

export default function Markdown({ children, className = '' }: MarkdownProps) {
  return (
    <div className={`prose prose-sm max-w-none ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[
          rehypeRaw,
          [rehypeSanitize, sanitizeSchema],
        ]}
        components={{
          a: ({ node, ...props }) => (
            <a
              {...props}
              target="_blank"
              rel="noopener noreferrer nofollow"
              className="text-gray-700 hover:text-gray-900 transition-colors duration-300"
            />
          ),
          code: ({ node, className, ...props }) => {
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
          pre: ({ node, ...props }) => (
            <pre
              {...props}
              className="bg-gray-50 p-4 rounded overflow-x-auto text-sm"
            />
          ),
          blockquote: ({ node, ...props }) => (
            <blockquote
              {...props}
              className="border-l-4 border-gray-300 pl-4 italic text-gray-600"
            />
          ),
          table: ({ node, ...props }) => (
            <div className="overflow-x-auto">
              <table {...props} className="min-w-full border-collapse" />
            </div>
          ),
          th: ({ node, ...props }) => (
            <th
              {...props}
              className="border border-gray-300 px-4 py-2 bg-gray-100 text-left font-semibold"
            />
          ),
          td: ({ node, ...props }) => (
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

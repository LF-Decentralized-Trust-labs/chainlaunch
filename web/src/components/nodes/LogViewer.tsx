import { useRef, useEffect } from 'react'
import Convert from 'ansi-to-html'
import DOMPurify from 'dompurify'

interface LogViewerProps {
	logs: string
	onScroll: (isScrolledToBottom: boolean) => void
	onLoadMore?: () => void
	isLoading?: boolean
	autoScroll?: boolean
}

const converter = new Convert({
	fg: '#fff',
	bg: '#000',
	newline: true,
	escapeXML: true,
})

const formatLogs = (logs: string) => {
	if (!logs) return ''
	const converted = converter.toHtml(logs)
	const sanitized = DOMPurify.sanitize(converted)
	return sanitized
}

export function LogViewer({ logs, onScroll, onLoadMore, isLoading = false, autoScroll = true }: LogViewerProps) {
	const logsRef = useRef<HTMLPreElement>(null)
	const lastScrollHeight = useRef<number>(0)
	const isUserScrolling = useRef<boolean>(false)

	useEffect(() => {
		if (!logsRef.current) return

		const logsElement = logsRef.current
		const { scrollHeight, clientHeight, scrollTop } = logsElement
		const isScrolledToBottom = scrollHeight - clientHeight <= scrollTop + 150

		// If we're loading more logs from the top
		if (lastScrollHeight.current) {
			const newScrollHeight = scrollHeight
			const scrollDiff = newScrollHeight - lastScrollHeight.current
			logsElement.scrollTop += scrollDiff
			lastScrollHeight.current = 0
			return
		}

		// Auto-scroll to bottom for new logs if user hasn't scrolled up
		if ((autoScroll && isScrolledToBottom) || !isUserScrolling.current) {
			logsElement.scrollTop = scrollHeight
		}
	}, [logs, autoScroll])

	const handleScroll = (e: React.UIEvent<HTMLPreElement>) => {
		const target = e.target as HTMLPreElement
		const isScrolledToBottom = target.scrollHeight - target.clientHeight <= target.scrollTop + 150

		// Update user scrolling state
		isUserScrolling.current = !isScrolledToBottom
		onScroll(isScrolledToBottom)

		if (target.scrollTop === 0 && onLoadMore) {
			lastScrollHeight.current = target.scrollHeight
			onLoadMore()
		}
	}

	return (
		<div className="relative">
			{isLoading && (
				<div className="absolute top-0 left-0 right-0 flex justify-center p-2 bg-background/80 backdrop-blur-sm">
					<div className="flex items-center gap-2 text-sm text-muted-foreground">
						<div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
						Loading more logs...
					</div>
				</div>
			)}
			<pre
				ref={logsRef}
				dangerouslySetInnerHTML={{ __html: formatLogs(logs) }}
				className="font-mono text-sm h-[600px] p-4 rounded-md overflow-auto bg-white dark:bg-black"
				style={{
					whiteSpace: 'pre-wrap',
					wordBreak: 'break-all',
				}}
				onScroll={handleScroll}
			/>
		</div>
	)
}

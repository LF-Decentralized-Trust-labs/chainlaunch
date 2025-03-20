import { format, formatDistanceToNow } from 'date-fns'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './tooltip'

interface TimeAgoProps {
	date: string | Date
	className?: string
}

export function TimeAgo({ date, className }: TimeAgoProps) {
	const dateObj = typeof date === 'string' ? new Date(date) : date
	const timeAgo = formatDistanceToNow(dateObj, { addSuffix: true })
	const fullDate = format(dateObj, 'PPpp')

	return (
		<TooltipProvider>
			<Tooltip>
				<TooltipTrigger asChild>
					<time dateTime={dateObj.toISOString()} className={className} title={fullDate}>
						{timeAgo}
					</time>
				</TooltipTrigger>
				<TooltipContent>
					<p>{fullDate}</p>
				</TooltipContent>
			</Tooltip>
		</TooltipProvider>
	)
}

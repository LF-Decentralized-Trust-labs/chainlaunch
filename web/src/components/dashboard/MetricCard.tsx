import { Card, CardContent } from '@/components/ui/card'

interface MetricCardProps {
	title: string
	value: string | number
	change: string
	icon?: React.ReactNode
}

export function MetricCard({ title, value, change, icon }: MetricCardProps) {
	return (
		<Card>
			<CardContent className="p-6">
				<div className="flex items-center justify-between">
					<div className="space-y-1">
						<p className="text-sm font-medium text-muted-foreground">{title}</p>
						<p className="text-2xl font-bold">{value}</p>
						<p className="text-sm text-muted-foreground">{change}</p>
					</div>
					{icon && <div className="text-muted-foreground">{icon}</div>}
				</div>
			</CardContent>
		</Card>
	)
}

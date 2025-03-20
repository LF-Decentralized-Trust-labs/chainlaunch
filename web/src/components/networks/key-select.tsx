import { ModelsKeyResponse } from '@/api/client'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'

interface KeySelectProps {
	keys: ModelsKeyResponse[]
	value: number[]
	onChange: (value: number[]) => void
}

export function KeySelect({ keys, value, onChange }: KeySelectProps) {
	return (
		<Card className="border-dashed">
			<ScrollArea className="h-[200px] p-4">
				<div className="space-y-4">
					{keys.map((key) => (
						<div key={key.id} className={cn('flex items-center space-x-4 rounded-md border p-4', value.includes(key.id!) && 'border-primary')}>
							<Checkbox
								checked={value.includes(key.id!)}
								onCheckedChange={(checked) => {
									if (checked) {
										onChange([...value, key.id!])
									} else {
										onChange(value.filter((id) => id !== key.id))
									}
								}}
							/>
							<div className="flex-1 space-y-1">
								<p className="text-sm font-medium leading-none">{key.name}</p>
								<p className="text-sm text-muted-foreground">Created {new Date(key.createdAt!).toLocaleDateString()}</p>
							</div>
							<Badge variant="outline">{key.algorithm}</Badge>
						</div>
					))}
				</div>
			</ScrollArea>
		</Card>
	)
}

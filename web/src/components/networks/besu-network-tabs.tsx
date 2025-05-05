import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export type BesuTabValue = 'details' | 'genesis'

interface BesuNetworkTabsProps {
	tab: BesuTabValue
	setTab: (tab: BesuTabValue) => void
	networkDetails: React.ReactNode
	genesis: React.ReactNode
}

export function BesuNetworkTabs({ tab, setTab, networkDetails, genesis }: BesuNetworkTabsProps) {
	return (
		<Tabs value={tab} onValueChange={(value) => setTab(value as BesuTabValue)}>
			<TabsList>
				<TabsTrigger value="details">Details</TabsTrigger>
				<TabsTrigger value="genesis">Genesis</TabsTrigger>
			</TabsList>

			<TabsContent className="mt-8" value="details">
				{networkDetails}
			</TabsContent>

			<TabsContent className="mt-8" value="genesis">
				{genesis}
			</TabsContent>
		</Tabs>
	)
} 
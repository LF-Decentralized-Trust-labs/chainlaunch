import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ProBadge } from '@/components/pro/ProBadge'
import { ProFeatureGate } from '@/components/pro/ProFeatureGate'
import { useState, useEffect } from 'react'

export type TabValue = 'details' | 'genesis' | 'anchor-peers' | 'consenters' | 'chaincode' | 'share' | 'channel-update' | 'proposals'

interface NetworkTabsProps {
	tab: TabValue
	setTab: (tab: TabValue) => void
	networkDetails: React.ReactNode
	anchorPeers?: React.ReactNode
	consenters?: React.ReactNode
	chaincode?: React.ReactNode
	share?: React.ReactNode
	channelUpdate?: React.ReactNode
	proposals?: React.ReactNode
}

export function NetworkTabs({ tab, setTab, networkDetails, anchorPeers, consenters, chaincode, share, channelUpdate, proposals }: NetworkTabsProps) {
	// Check if current tab is a pro feature and redirect to details if needed
	useEffect(() => {
		const proTabs: TabValue[] = ['share', 'channel-update', 'proposals'];
		if (proTabs.includes(tab)) {
			setTab('details');
		}
	}, [tab, setTab]);

	// Handle tab change with pro feature check
	const handleTabChange = (value: string) => {
		const proTabs: TabValue[] = ['share', 'channel-update', 'proposals'];
		
		if (proTabs.includes(value as TabValue)) {
			// Don't change tab, it will be handled by the TabsTrigger onClick
			return;
		}
		
		setTab(value as TabValue);
	};

	// Render pro feature gate for specific tabs
	const renderProContent = (title: string, description: string) => {
		return (
			<ProFeatureGate
				title={title}
				description={description}
			/>
		);
	};

	return (
		<Tabs value={tab} onValueChange={handleTabChange}>
			<TabsList>
				<TabsTrigger value="details">Details</TabsTrigger>
				{anchorPeers && <TabsTrigger value="anchor-peers">Anchor Peers</TabsTrigger>}
				{consenters && <TabsTrigger value="consenters">Consenters</TabsTrigger>}
				{chaincode && <TabsTrigger value="chaincode">Chaincode</TabsTrigger>}
				
				{channelUpdate && (
					<TabsTrigger 
						value="channel-update" 
						onClick={() => window.open("https://chainlaunch.dev/premium", "_blank")}
						className="flex items-center gap-2"
					>
						Channel Update
						<ProBadge />
					</TabsTrigger>
				)}
				
				{proposals && (
					<TabsTrigger 
						value="proposals" 
						onClick={() => window.open("https://chainlaunch.dev/premium", "_blank")}
						className="flex items-center gap-2"
					>
						Proposals
						<ProBadge />
					</TabsTrigger>
				)}
				
				{share && (
					<TabsTrigger 
						value="share" 
						onClick={() => window.open("https://chainlaunch.dev/premium", "_blank")}
						className="flex items-center gap-2"
					>
						Share
						<ProBadge />
					</TabsTrigger>
				)}
			</TabsList>
			
			<TabsContent className="mt-8" value="details">
				{networkDetails}
			</TabsContent>
			
			{anchorPeers && <TabsContent className="mt-8" value="anchor-peers">{anchorPeers}</TabsContent>}
			{consenters && <TabsContent className="mt-8" value="consenters">{consenters}</TabsContent>}
			{chaincode && <TabsContent className="mt-8" value="chaincode">{chaincode}</TabsContent>}
			
			{channelUpdate && (
				<TabsContent className="mt-8" value="channel-update">
					{renderProContent(
						"Channel Update Pro Feature",
						"Upgrade to ChainLaunch Pro to access advanced channel update capabilities, enabling seamless network configuration changes."
					)}
				</TabsContent>
			)}
			
			{proposals && (
				<TabsContent className="mt-8" value="proposals">
					{renderProContent(
						"Proposals Pro Feature",
						"Upgrade to ChainLaunch Pro to manage network proposals, enabling collaborative governance and decision-making across organizations."
					)}
				</TabsContent>
			)}
			
			{share && (
				<TabsContent className="mt-8" value="share">
					{renderProContent(
						"Network Sharing Pro Feature",
						"Upgrade to ChainLaunch Pro to share networks with other organizations, enabling cross-organizational collaboration and network participation."
					)}
				</TabsContent>
			)}
		</Tabs>
	)
}

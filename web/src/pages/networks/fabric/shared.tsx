import { ProFeatureGate } from '@/components/pro/ProFeatureGate'

export default function SharedNetworksPage() {
	return (
		<ProFeatureGate
			title="Access Shared Networks"
			description="Upgrade to ChainLaunch Pro to access networks shared by other organizations, enabling cross-organizational collaboration and network participation."
		/>
	)
}

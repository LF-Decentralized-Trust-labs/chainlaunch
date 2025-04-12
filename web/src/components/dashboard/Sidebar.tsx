import { Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupLabel, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '@/components/ui/sidebar'
import { BadgeCheck, Bell, Building, ChevronsUpDown, DatabaseBackup, Globe, Key, LogOut, Network, Server, Share2, Settings } from 'lucide-react'
;('use client')

// import { Project } from '@/api/client'
import { useAuth } from '@/contexts/AuthContext'
// import { useProjects } from '@/contexts/ProjectsContext'
import { type LucideIcon } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'
import logo from '../../../public/logo.svg'
import { Avatar, AvatarFallback } from '../ui/avatar'
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '../ui/dropdown-menu'
import { ProBadge } from '../pro/ProBadge'

type NavItem = {
	title: string
	url: string
	icon: LucideIcon | React.FC<{ className?: string }>
	isPro?: boolean
}

const data = {
	navMain: [
		{
			title: 'Platform',
			items: [
				{
					title: 'Nodes',
					url: '/nodes',
					icon: Server,
				},
				{
					title: 'Monitoring',
					url: '/settings/monitoring',
					icon: Bell,
				},
				{
					title: 'Networks',
					url: '/networks',
					icon: Network,
				},
			],
		},
		{
			title: 'Fabric',
			items: [
				{
					title: 'Organizations',
					url: '/fabric/organizations',
					icon: Building,
				},
			],
		},
		{
			title: 'Settings',
			items: [
				{
					title: 'Key management',
					url: '/settings/keys',
					icon: Key,
				},
				{
					title: 'Backups',
					url: '/settings/backups',
					icon: DatabaseBackup,
				},
				{
					title: 'Settings',
					url: '/settings/general',
					icon: Settings,
				},
			],
		},
		{
			title: 'Connect',
			items: [
				{
					title: 'Connect',
					url: '/connect',
					icon: Globe,
					isPro: true,
				},
				{
					title: 'External Nodes',
					url: '/external-nodes',
					icon: Network,
					isPro: true,
				},
				{
					title: 'Shared Networks',
					url: '/networks/fabric/shared',
					icon: Share2,
					isPro: true,
				},
			],
		},
		// {
		// 	title: 'Decentralized Identity',
		// 	items: [
		// 		{
		// 			title: 'Issuers',
		// 			url: '/identity/issuers',
		// 			icon: BadgeCheck,
		// 		},
		// 		{
		// 			title: 'Verifiers',
		// 			url: '/identity/verifiers',
		// 			icon: ShieldCheck,
		// 		},
		// 		{
		// 			title: 'Certificate templates',
		// 			url: '/identity/certificates',
		// 			icon: FileText,
		// 		},
		// 	],
		// },
	],
}

function NavMain({
	items,
}: {
	items: {
		title: string
		items: NavItem[]
	}[]
}) {
	const location = useLocation()

	return (
		<>
			{items.map((section) => (
				<SidebarGroup key={section.title}>
					<SidebarGroupLabel>{section.title}</SidebarGroupLabel>
					<SidebarMenu>
						{section.items.map((item) => {
							const isActive = location.pathname.startsWith(item.url)
							const Icon = item.icon
							return (
								<SidebarMenuItem key={item.title}>
									<SidebarMenuButton asChild tooltip={item.title} className={isActive ? 'bg-sidebar-accent text-sidebar-accent-foreground' : ''}>
										<Link to={item.url}>
											<Icon className="size-4" />
											<span>{item.title}</span>
											{item.isPro && <ProBadge />}
										</Link>
									</SidebarMenuButton>
								</SidebarMenuItem>
							)
						})}
					</SidebarMenu>
				</SidebarGroup>
			))}
		</>
	)
}

export default function AppSidebar() {
	return (
		<Sidebar>
			<SidebarHeader>
				<SidebarMenu>
					<SidebarMenuItem>
						<div className="flex items-center gap-2">
							<div className="flex aspect-square size-8 items-center justify-center rounded-lg  dark:text-sidebar-primary-foreground bg-black dark:bg-transparent">
								<img src={logo} alt="logo" className="size-full" />
							</div>
							<div className="grid flex-1 text-left text-sm leading-tight">
								<span className="truncate font-semibold">ChainLaunch</span>
								<span className="truncate text-xs">v1.0.0</span>
							</div>
						</div>
					</SidebarMenuItem>
				</SidebarMenu>
			</SidebarHeader>
			<SidebarContent>
				<NavMain items={data.navMain} />
			</SidebarContent>
			<SidebarFooter>
				<NavUser />
			</SidebarFooter>
		</Sidebar>
	)
}

function NavUser() {
	const { user } = useAuth()
	const { isMobile } = useSidebar()
	const { logout } = useAuth()
	if (!user) return null

	return (
		<SidebarMenu>
			<SidebarMenuItem>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<SidebarMenuButton size="lg" className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground">
							<Avatar className="h-8 w-8 rounded-lg">
								<AvatarFallback className="rounded-lg">{user.username?.slice(0, 2).toUpperCase() || 'U'}</AvatarFallback>
							</Avatar>
							<div className="grid flex-1 text-left text-sm leading-tight">
								<span className="truncate font-semibold">{user.username || 'User'}</span>
							</div>
							<ChevronsUpDown className="ml-auto size-4" />
						</SidebarMenuButton>
					</DropdownMenuTrigger>
					<DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg" side={isMobile ? 'bottom' : 'right'} align="end" sideOffset={4}>
						<DropdownMenuLabel className="p-0 font-normal">
							<div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
								<Avatar className="h-8 w-8 rounded-lg">
									<AvatarFallback className="rounded-lg">{user.username?.slice(0, 2).toUpperCase() || 'U'}</AvatarFallback>
								</Avatar>
								<div className="grid flex-1 text-left text-sm leading-tight">
									<span className="truncate font-semibold">{user.username || 'User'}</span>
								</div>
							</div>
						</DropdownMenuLabel>
						<DropdownMenuSeparator />
						{/* 
						<DropdownMenuGroup>
							<DropdownMenuItem>
								<BadgeCheck />
								Account
							</DropdownMenuItem>
						</DropdownMenuGroup> */}
						<DropdownMenuSeparator />
						<DropdownMenuItem
							onClick={async () => {
								await logout()
								// await logoutMutation({})
								// location.reload()
							}}
						>
							<LogOut />
							Log out
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</SidebarMenuItem>
		</SidebarMenu>
	)
}

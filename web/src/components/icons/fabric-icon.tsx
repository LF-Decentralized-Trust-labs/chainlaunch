import fabricLogo from '../../../public/blockchains/fabric_stroke.svg'

interface FabricIconProps {
  className?: string
}

export function FabricIcon({ className }: FabricIconProps) {
  return <img src={fabricLogo} alt="Fabric" className={`[&_*]:fill-black dark:[&_*]:fill-white ${className}`} />
} 
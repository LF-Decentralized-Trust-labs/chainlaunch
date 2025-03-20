import { Outlet } from 'react-router-dom';
import AppSidebar from '@/components/dashboard/Sidebar';

const DashboardLayout = () => {
  return (
    <div className="flex min-h-screen">
      <AppSidebar />
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  );
};

export default DashboardLayout; 
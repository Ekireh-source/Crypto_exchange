'use client';

import { useState, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useSelector, useDispatch } from "react-redux";
import { logout, selectUser } from "@/store/authSlice";
import { Icon } from "@iconify/react";
import Image from "next/image";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

interface IUser {
  firstname?: string;
  lastname?: string;
  email?: string;
}

export default function FloatingNavbar() {
  const router = useRouter();
  const pathname = usePathname();
  const dispatch = useDispatch();
 

  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 10);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const handleLogout = () => {
    dispatch(logout());
    router.push("/login");
  };

  const currentUser = useSelector(selectUser) as IUser;

  const ALL_LINKS = [
    { label: "Home", icon: "hugeicons:home-03", href: "/dashboard" },
    { label: "P2P", icon: "hugeicons:trade-up", href: "/p2p" },
    { label: "Trade", icon: "hugeicons:chart-line-up-01", href: "/trade" },
    { label: "Swap", icon: "hugeicons:arrow-turn-backward", href: "/swap" },
    { label: "Transactions", icon: "hugeicons:bitcoin-transaction", href: "/transactions" },
    { label: "See more", icon: "hugeicons:more", href: "/more" },
  ];

  return (
    <nav
      className={cn(
        "w-full max-w-[1440px] mx-auto px-4 sm:px-6 lg:px-8 pt-6 pb-2 sticky top-0 z-[50] transition-all duration-300 pointer-events-none"
      )}
    >
      <div
        className={cn(
          "pointer-events-auto flex items-center justify-between rounded-[24px] px-6 h-[72px] transition-all duration-300 w-full",
          "bg-[#13151a] border border-[#22252e] shadow-[0_8px_30px_rgb(0,0,0,0.6)]",
          scrolled && "bg-[#13151a]/80 backdrop-blur-xl border-[#2a2d36] shadow-[0_8px_30px_rgba(37,99,235,0.15)]"
        )}
      >
        {/* Left Side: Logo + Nav Links */}
        <div className="flex items-center gap-6 h-full flex-1">
          {/* Logo */}
          <div
            onClick={() => router.push("/dashboard")}
            className="cursor-pointer flex items-center gap-3 shrink-0"
          >
            <div className="relative size-[32px] flex items-center justify-center bg-white rounded-full p-1">
              {/* Using a placeholder for the C logo from the image, or Coinbase style */}
              <span className="text-[#0a0b0d] font-bold text-xl leading-none">R</span>
            </div>
          </div>

          {/* Nav Links */}
          <ul className="hidden xl:flex items-center gap-6 list-none m-0 p-0 h-full ml-4">
            {ALL_LINKS.map((link) => {
              const isActive = pathname === link.href || (link.href !== "/dashboard" && pathname?.startsWith(link.href));
              return (
                <li key={link.href} className="relative flex items-center">
                  <span
                    onClick={() => router.push(link.href)}
                    className={cn(
                      "cursor-pointer text-[15px] font-semibold transition-all duration-200 px-1 whitespace-nowrap flex items-center gap-2",
                      isActive ? "text-white" : "text-[#888c99] hover:text-white"
                    )}
                  >
                    <Icon icon={link.icon} className="size-5 hidden lg:block" />
                    {link.label}
                  </span>
                  {isActive && (
                    <div className="absolute -bottom-2 left-0 w-full h-[3px] bg-blue-600 rounded-full" />
                  )}
                </li>
              );
            })}
          </ul>
        </div>

        {/* Right Side Actions */}
        <div className="flex items-center gap-4">
          <button className="hidden sm:flex items-center justify-center size-9 rounded-full bg-[#16181d] text-[#888c99] hover:text-white border border-[#22252e] transition-colors">
            <Icon icon="hugeicons:notification-01" className="size-5" />
          </button>
          
          <button className="hidden sm:flex items-center justify-center size-9 rounded-full bg-[#16181d] text-[#888c99] hover:text-white border border-[#22252e] transition-colors">
            <Icon icon="hugeicons:help-circle" className="size-5" />
          </button>
          
          <button className="hidden sm:flex items-center justify-center size-9 rounded-full bg-[#16181d] text-[#888c99] hover:text-white border border-[#22252e] transition-colors">
            <Icon icon="hugeicons:grid-view" className="size-5" />
          </button>

          {/* Profile Section */}
          <DropdownMenu>
            <DropdownMenuTrigger>
              <div className="flex items-center rounded-full cursor-pointer hover:opacity-80 transition-opacity">
                <Avatar className="size-9 border border-[#22252e]">
                  <AvatarImage src="/images/profile-placeholder.jpg" />
                  <AvatarFallback className="bg-blue-600 text-white text-sm font-semibold">
                    {currentUser?.firstname?.[0] || 'A'}
                  </AvatarFallback>
                </Avatar>
              </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[220px] mt-2 rounded-[12px] p-2 bg-[#16181d] border-[#22252e] shadow-xl">
              <DropdownMenuItem className="!items-start rounded-lg hover:bg-[#22252e] focus:bg-[#22252e] cursor-pointer transition-all mb-2 px-3 py-3">
                <div className="flex items-center justify-start gap-3">
                  <Avatar className="size-10">
                    <AvatarFallback className="bg-blue-600 text-white text-sm font-bold">
                      {currentUser?.firstname?.[0] || 'A'}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col min-w-0">
                    <span className="text-[15px] font-semibold text-white truncate leading-tight">
                      {currentUser?.firstname || 'User'} {currentUser?.lastname || ''}
                    </span>
                    <span className="text-[13px] text-[#888c99] truncate">
                      {currentUser?.email || 'user@example.com'}
                    </span>
                  </div>
                </div>
              </DropdownMenuItem>

              <DropdownMenuSeparator className="bg-[#22252e] mx-1" />
               
              <DropdownMenuItem
                onClick={() => router.push("/profile")}
                className="rounded-lg h-11 gap-3 font-medium text-[#e2e4e9] hover:bg-[#22252e] focus:bg-[#22252e] cursor-pointer px-3 mt-1"
              >
                <Icon icon="hugeicons:user" className="size-5 text-[#888c99]" />
                <span>Profile</span>
              </DropdownMenuItem>

              <DropdownMenuItem
                onClick={() => router.push("/settings")}
                className="rounded-lg h-11 gap-3 font-medium text-[#e2e4e9] hover:bg-[#22252e] focus:bg-[#22252e] cursor-pointer px-3 mt-1"
              >
                <Icon icon="hugeicons:settings-02" className="size-5 text-[#888c99]" />
                <span>Settings</span>
              </DropdownMenuItem>

              <DropdownMenuSeparator className="bg-[#22252e] mx-1" />

              <DropdownMenuItem
                onClick={handleLogout}
                className="rounded-lg h-11 gap-3 font-medium text-red-500 hover:bg-red-500/10 focus:bg-red-500/10 cursor-pointer px-3 mt-1"
              >
                <Icon icon="hugeicons:logout-02" className="size-5" />
                <span>Sign out</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Mobile menu burger */}
          <button
            onClick={() => setMenuOpen((v) => !v)}
            className="xl:hidden flex h-10 w-10 items-center justify-center rounded-full bg-[#16181d] text-[#e2e4e9] border border-[#22252e]"
          >
            <Icon icon={menuOpen ? "hugeicons:cancel-01" : "hugeicons:menu-11"} className="size-6" />
          </button>
        </div>
      </div>

      {/* Mobile Dropdown */}
      {menuOpen && (
        <div className="absolute left-4 right-4 mt-3 origin-top xl:hidden z-[60] rounded-[16px] border border-[#22252e] bg-[#16181d] shadow-2xl p-3 animate-in slide-in-from-top-2 fade-in duration-200">
          <div className="flex flex-col gap-1">
            {ALL_LINKS.map((link) => {
              const isActive = pathname === link.href || (link.href !== "/dashboard" && pathname?.startsWith(link.href));
              return (
                <div
                  key={link.href}
                  onClick={() => {
                    setMenuOpen(false);
                    router.push(link.href);
                  }}
                  className={cn(
                    "cursor-pointer rounded-xl px-4 py-3 flex items-center gap-3 transition-colors",
                    isActive ? "bg-blue-600/10 text-blue-500" : "text-[#e2e4e9] hover:bg-[#22252e]"
                  )}
                >
                  <Icon icon={link.icon} className="size-5" />
                  <span className="text-[15px] font-semibold">{link.label}</span>
                </div>
              );
            })}
            <div className="h-[1px] bg-[#22252e] my-2 mx-2" />
            <div
              onClick={() => {
                setMenuOpen(false);
                router.push("/profile");
              }}
              className="cursor-pointer rounded-xl px-4 py-3 text-[15px] font-semibold text-[#e2e4e9] hover:bg-[#22252e] transition-colors flex items-center gap-3"
            >
              <Icon icon="hugeicons:user" className="size-5" />
              Profile
            </div>
            <div
              onClick={() => {
                setMenuOpen(false);
                handleLogout();
              }}
              className="cursor-pointer rounded-xl px-4 py-3 text-[15px] font-semibold text-red-500 hover:bg-red-500/10 transition-colors flex items-center gap-3"
            >
              <Icon icon="hugeicons:logout-02" className="size-5" />
              Sign out
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}
import Link from "next/link";
import { ArrowDown, Cpu, Network, Database, Zap } from "lucide-react";

export default function Home() {
  return (
    <div className="min-h-screen bg-[#01010d] text-slate-200 overflow-hidden font-sans relative selection:bg-pink-500/30 selection:text-pink-100">
      {/* Background Grid */}
      <div className="absolute inset-0 bg-grid-pattern opacity-40 pointer-events-none"></div>

      {/* Top Gradient */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[500px] bg-indigo-900/20 blur-[120px] rounded-full pointer-events-none"></div>

      {/* Header */}
      <header className="fixed top-4 md:top-6 left-1/2 -translate-x-1/2 w-[95%] max-w-6xl z-50 flex items-center justify-between px-4 md:px-6 py-3 md:py-4 border border-white/10 rounded-full bg-[#01010d]/70 backdrop-blur-md shadow-2xl">
        <div className="flex items-center gap-3">
          <div className="w-5 h-5 border-[3px] border-indigo-400 rounded-sm"></div>
          <span className="text-xl font-medium tracking-tight text-white">BusyWT</span>
        </div>
        
        <nav className="hidden md:flex items-center gap-10 text-sm text-slate-400">
          <Link href="#market" className="hover:text-white transition-colors">P2P Market</Link>
          <Link href="#wallets" className="hover:text-white transition-colors">Wallets</Link>
          <Link href="#security" className="hover:text-white transition-colors">Security</Link>
          <Link href="#support" className="hover:text-white transition-colors">Support</Link>
        </nav>

        <div className="flex items-center gap-4 text-sm font-medium">
          <Link href="/login" className="px-5 py-2 border border-white/10 rounded-full hover:bg-white/5 transition-colors">Log In</Link>
          <Link href="/signup" className="px-5 py-2 border border-indigo-500/50 bg-indigo-500/10 rounded-full hover:bg-indigo-500/20 transition-colors">Sign Up</Link>
        </div>
      </header>

      {/* Main Content */}
      <main className="relative z-10 flex flex-col items-center pt-32 pb-32 px-4 text-center">
        
        {/* Core Visualization */}
        <div className="relative mb-24 w-full max-w-4xl mx-auto flex items-center justify-center">
          
          {/* Central Unit */}
          <div className="relative z-20 border border-white/10 bg-[#01010d]/80 backdrop-blur-md p-6 rounded-lg shadow-[0_0_50px_rgba(99,102,241,0.1)]">
             <div className="absolute -top-3 left-1/2 -translate-x-1/2 text-[10px] text-pink-400 tracking-[0.2em] font-mono whitespace-nowrap bg-[#01010d] px-3 py-1 border border-pink-500/30 rounded-sm">
               BUSYWT AI CORE 15.1
             </div>
             
             {/* Crosshairs on corners */}
             <div className="absolute -top-1 -left-1 w-2 h-2 border-t border-l border-white/30"></div>
             <div className="absolute -top-1 -right-1 w-2 h-2 border-t border-r border-white/30"></div>
             <div className="absolute -bottom-1 -left-1 w-2 h-2 border-b border-l border-white/30"></div>
             <div className="absolute -bottom-1 -right-1 w-2 h-2 border-b border-r border-white/30"></div>

             <div className="flex flex-col md:flex-row items-center gap-8 md:gap-12 p-6 md:p-10 border border-white/5 rounded-md relative overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-b from-indigo-500/5 to-transparent"></div>

                {/* Left metrics */}
                <div className="flex flex-row md:flex-col gap-4 md:gap-10 w-full md:w-auto justify-around">
                  <div className="flex flex-col items-center md:items-end gap-1">
                    <span className="text-[10px] text-slate-500 font-mono tracking-widest">NEURAL</span>
                    <span className="text-[11px] text-slate-300 font-medium tracking-widest">INTELLIGENCE</span>
                  </div>
                  <div className="flex flex-col items-center md:items-end gap-1">
                    <span className="text-[10px] text-slate-500 font-mono tracking-widest">EXECUTION</span>
                    <span className="text-[11px] text-slate-300 font-medium tracking-widest">ENGINE</span>
                  </div>
                </div>

                {/* Center Core */}
                <div className="relative w-32 h-32 flex items-center justify-center border border-indigo-500/30 rounded-lg neon-border bg-indigo-950/30 crosshairs animate-pulse-blue">
                  <Cpu className="w-12 h-12 text-indigo-300 opacity-80" />
                  <div className="absolute inset-0 border border-indigo-400/20 rounded-lg rotate-45 scale-[0.8] mix-blend-screen animate-spin-slow"></div>
                  <div className="absolute inset-0 border border-indigo-400/20 rounded-lg -rotate-45 scale-[0.8] mix-blend-screen animate-spin-slow-reverse"></div>
                </div>

                {/* Right metrics */}
                <div className="flex flex-row md:flex-col gap-4 md:gap-10 w-full md:w-auto justify-around">
                  <div className="flex flex-col items-center md:items-start gap-1">
                    <span className="text-[10px] text-slate-500 font-mono tracking-widest">NEURAL</span>
                    <span className="text-[11px] text-slate-300 font-medium tracking-widest">NETWORKS</span>
                  </div>
                  <div className="flex flex-col items-center md:items-start gap-1">
                    <span className="text-[10px] text-slate-500 font-mono tracking-widest">DATA</span>
                    <span className="text-[11px] text-slate-300 font-medium tracking-widest">MASTERY</span>
                  </div>
                </div>
             </div>
          </div>

          {/* Connection Lines & Nodes */}
          {/* BTC Node */}
          <div className="absolute left-0 top-1/2 -translate-y-1/2 flex items-center gap-4 hidden md:flex animate-float">
             <div className="w-24 h-24 border border-pink-500/30 neon-border-pink rounded-md bg-pink-950/30 flex flex-col items-center justify-center gap-2 crosshairs z-10 animate-pulse-pink">
               <span className="text-[10px] font-mono text-pink-300 tracking-widest">BITCOIN</span>
               <div className="text-pink-400/60"><Network size={20} strokeWidth={1.5} /></div>
             </div>
             {/* Glowing Line */}
             <div className="h-0.5 w-32 bg-pink-900/50 relative overflow-hidden">
               <div className="absolute inset-0 bg-pink-500 blur-[2px] opacity-30"></div>
               <div className="flow-line-right"></div>
             </div>
          </div>

          {/* ETH Node */}
          <div className="absolute right-0 top-1/2 -translate-y-1/2 flex items-center gap-4 hidden md:flex animate-float" style={{ animationDelay: '1s' }}>
             {/* Glowing Line */}
             <div className="h-0.5 w-32 bg-indigo-900/50 relative overflow-hidden">
               <div className="absolute inset-0 bg-indigo-500 blur-[2px] opacity-30"></div>
               <div className="flow-line-left"></div>
             </div>
             <div className="w-24 h-24 border border-indigo-500/30 neon-border rounded-md bg-indigo-950/30 flex flex-col items-center justify-center gap-2 crosshairs z-10 animate-pulse-blue" style={{ animationDelay: '1s' }}>
               <span className="text-[10px] font-mono text-indigo-300 tracking-widest">ETHEREUM</span>
               <div className="text-indigo-400/60"><Database size={20} strokeWidth={1.5} /></div>
             </div>
          </div>

          {/* Top Sell/Profit Tag */}
          <div className="absolute -top-16 left-1/2 -translate-x-1/2 flex flex-col items-center">
            <div className="border border-pink-500/30 bg-pink-500/10 px-4 py-1.5 rounded-sm flex items-center gap-2 shadow-[0_0_15px_rgba(244,114,182,0.15)]">
              <span className="text-[11px] text-slate-400 font-mono tracking-widest">SELL BTC:</span>
              <span className="text-[11px] text-pink-400 font-mono tracking-widest">[ PROFIT <span className="text-pink-300">-1.4%</span> ]</span>
            </div>
            <ArrowDown className="text-pink-500 w-4 h-4 mt-2 opacity-80" strokeWidth={1.5} />
          </div>
          
        </div>

        {/* Hero Copy */}
        <div className="relative inline-block mt-4 mb-8">
          <div className="absolute -top-4 -left-4 w-3 h-3 border-t border-l border-white/20"></div>
          <div className="absolute -top-4 -right-4 w-3 h-3 border-t border-r border-white/20"></div>
          <div className="absolute -bottom-4 -left-4 w-3 h-3 border-b border-l border-white/20"></div>
          <div className="absolute -bottom-4 -right-4 w-3 h-3 border-b border-r border-white/20"></div>
          
          <h1 className="text-4xl sm:text-5xl md:text-7xl font-semibold tracking-tight text-white px-4 md:px-8 py-4 glow-text">
            The ultimate P2P<br/>crypto exchange
          </h1>
        </div>
        
        <p className="max-w-2xl text-slate-400 text-lg md:text-xl leading-relaxed mb-12">
          BusyWT is the premier P2P crypto trading platform &mdash; secure, lightning-fast,
          and giving you full control over your digital assets.
        </p>
        
        <Link 
          href="/waitlist" 
          className="group relative px-8 py-4 bg-transparent border border-white/20 text-white font-medium hover:border-white/50 transition-all duration-300 overflow-hidden rounded-sm"
        >
          <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/20 to-pink-500/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
          <span className="relative z-10 flex items-center gap-2">
            Join the Waiting List
            <Zap className="w-4 h-4 text-indigo-400 group-hover:text-pink-400 transition-colors duration-300" />
          </span>
        </Link>
      </main>
    </div>
  );
}

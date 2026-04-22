import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { toast } from 'react-hot-toast';
import { Key, Eye, EyeOff, ArrowRight } from 'lucide-react';
import { AuthService } from '../services/api';
import { useAuthStore } from '../store/authStore';

const GithubIcon = ({ size = 16 }: { size?: number }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
    <path d="M9 18c-4.51 2-5-2-7-2" />
  </svg>
);

export const Register: React.FC = () => {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const { register, handleSubmit, formState: { errors } } = useForm({
    defaultValues: { first_name: '', last_name: '', email: '', password: '' }
  });

  const onSubmit = async (data: any) => {
    try {
      setIsLoading(true);
      const res = await AuthService.register(data);
      setAuth(res.data.token, res.data.user);
      toast.success('Account created successfully');
      navigate('/wallet');
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Registration failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] bg-white flex flex-col font-sans overflow-y-auto selection:bg-[#B5FF45]">
      {/* Background Grid */}
      <div 
        className="fixed inset-0 pointer-events-none opacity-[0.4]"
        style={{ 
          backgroundImage: `linear-gradient(#F0F0F0 1px, transparent 1px), linear-gradient(90deg, #F0F0F0 1px, transparent 1px)`,
          backgroundSize: '40px 40px'
        }}
      />

      {/* Navbar */}
      <header className="relative z-10 w-full px-[48px] h-[80px] flex items-center justify-between border-b border-[#F0F0F0] bg-white/80 backdrop-blur-md">
        <Link to="/" className="text-[20px] font-black tracking-tight cursor-pointer">TRADEX</Link>
        
        <nav className="flex items-center gap-[32px]">
          <Link to="/" className="text-[14px] text-[#666666] hover:text-black transition-colors">Home</Link>
          <Link to="/login" className="text-[14px] text-[#666666] hover:text-black transition-colors">Login</Link>
          <Link to="/register" className="text-[14px] text-black font-bold relative pb-1">
            Register
            <div className="absolute bottom-0 left-0 w-full h-[2px] bg-[#B5FF45]" />
          </Link>
        </nav>

        <Link to="/trade" className="h-[36px] px-[20px] rounded-full bg-[#B5FF45] text-[13px] font-bold flex items-center hover:brightness-105 transition-all shadow-sm">
          Trade Now
        </Link>
      </header>

      {/* Main Content */}
      <main className="relative z-10 flex-1 flex justify-center items-center py-[60px]">
        <div className="w-[480px] bg-white p-[48px] rounded-[16px] shadow-[0_20px_50px_rgba(0,0,0,0.04)] border border-[#F0F0F0]">
          
          <div className="mb-[40px]">
            <p className="text-[11px] font-bold tracking-[0.15em] text-[#999999] uppercase mb-[12px]">Registration Protocol</p>
            <h1 className="text-[36px] font-extrabold text-black leading-[1.1] mb-[12px]">
              Create your <br /> TRADEX account.
            </h1>
            <p className="text-[14px] text-[#666666]">Join the next generation of global market access.</p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-[24px]">
            <div className="grid grid-cols-2 gap-[16px]">
              <div>
                <label className="block text-[10px] font-bold text-[#999999] uppercase mb-[8px] tracking-wider">First Name</label>
                <input 
                  {...register('first_name', { required: 'Required' })}
                  type="text" 
                  placeholder="John"
                  className={`w-full h-[48px] bg-[#F3F3F3] rounded-[8px] px-[16px] text-[14px] outline-none focus:ring-2 focus:ring-[#B5FF45]/50 transition-all ${errors.first_name ? 'ring-2 ring-red-500/50' : ''}`}
                />
              </div>
              <div>
                <label className="block text-[10px] font-bold text-[#999999] uppercase mb-[8px] tracking-wider">Last Name</label>
                <input 
                  {...register('last_name', { required: 'Required' })}
                  type="text" 
                  placeholder="Doe"
                  className={`w-full h-[48px] bg-[#F3F3F3] rounded-[8px] px-[16px] text-[14px] outline-none focus:ring-2 focus:ring-[#B5FF45]/50 transition-all ${errors.last_name ? 'ring-2 ring-red-500/50' : ''}`}
                />
              </div>
            </div>

            <div>
              <label className="block text-[10px] font-bold text-[#999999] uppercase mb-[8px] tracking-wider">Email Identifier</label>
              <input 
                {...register('email', { required: 'Email is required' })}
                type="email" 
                placeholder="user@network.tradex"
                className={`w-full h-[48px] bg-[#F3F3F3] rounded-[8px] px-[16px] text-[14px] outline-none focus:ring-2 focus:ring-[#B5FF45]/50 transition-all ${errors.email ? 'ring-2 ring-red-500/50' : ''}`}
              />
            </div>

            <div className="relative">
              <label className="block text-[10px] font-bold text-[#999999] uppercase mb-[8px] tracking-wider">Security Protocol</label>
              <div className="relative">
                <input 
                  {...register('password', { required: 'Password is required', minLength: { value: 6, message: 'Min 6 chars' } })}
                  type={showPassword ? "text" : "password"} 
                  placeholder="••••••••••••••••"
                  className={`w-full h-[48px] bg-[#F3F3F3] rounded-[8px] px-[16px] pr-[44px] text-[14px] outline-none focus:ring-2 focus:ring-[#B5FF45]/50 transition-all ${errors.password ? 'ring-2 ring-red-500/50' : ''}`}
                />
                <button 
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-[16px] top-1/2 -translate-y-1/2 text-[#999999] hover:text-black"
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>

            <label className="flex items-center gap-[10px] cursor-pointer group">
              <div className="w-[18px] h-[18px] border-[2px] border-[#EEEEEE] rounded-[4px] flex items-center justify-center bg-[#F3F3F3] group-hover:border-[#B5FF45]">
                <input type="checkbox" className="hidden peer" />
                <div className="w-[10px] h-[10px] bg-[#B5FF45] rounded-[2px] opacity-0 peer-checked:opacity-100" />
              </div>
              <span className="text-[12px] text-[#666666] select-none">Accept Terms of Service Protocol</span>
            </label>

            <button 
              type="submit"
              disabled={isLoading}
              className="group relative w-full h-[54px] bg-[#B5FF45] rounded-[8px] font-bold text-[15px] flex items-center justify-center gap-[12px] transition-all active:translate-y-[2px] shadow-[0_4px_0_0_#9FD63D] hover:shadow-[0_2px_0_0_#9FD63D] disabled:opacity-70"
            >
              {isLoading ? 'Creating Account...' : 'Initialize Account'}
              <ArrowRight size={18} className="group-hover:translate-x-[4px] transition-transform" />
            </button>

            <div className="relative py-[20px] flex items-center justify-center">
              <div className="absolute w-full h-[1px] bg-[#F0F0F0]" />
              <span className="relative px-[16px] bg-white text-[10px] font-bold text-[#999999] uppercase tracking-widest">Or Register With</span>
            </div>

            <div className="grid grid-cols-2 gap-[16px]">
              <button type="button" className="h-[44px] bg-[#EEEEEE] rounded-[8px] flex items-center justify-center gap-[10px] text-[13px] font-bold hover:bg-[#E5E5E5] transition-colors">
                <GithubIcon size={16} /> Github
              </button>
              <button type="button" className="h-[44px] bg-[#EEEEEE] rounded-[8px] flex items-center justify-center gap-[10px] text-[13px] font-bold hover:bg-[#E5E5E5] transition-colors">
                <Key size={16} /> Passkey
              </button>
            </div>
          </form>
        </div>
      </main>

      {/* Footer */}
      <footer className="w-full px-[48px] py-[40px] flex flex-col md:flex-row justify-between items-center border-t border-[#F0F0F0] gap-4 bg-white relative z-10">
        <div>
          <div className="text-[16px] font-black mb-[8px]">TRADEX</div>
          <p className="text-[10px] text-[#999999] uppercase tracking-wider">
            © 2024 TRADEX. ALL RIGHTS RESERVED. BUILT FOR KINETIC VELOCITY.
          </p>
        </div>
        <div className="flex gap-[24px]">
          {['Terms of Service', 'Privacy Policy', 'Risk Disclosure', 'Help Center'].map((item) => (
            <a key={item} href="#" className="text-[10px] font-bold text-[#999999] uppercase hover:text-black transition-colors tracking-tighter">
              {item}
            </a>
          ))}
        </div>
      </footer>
    </div>
  );
};

import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import mkcert from 'vite-plugin-mkcert';

export default defineConfig({
	plugins: [ sveltekit(),mkcert() ],
	server: { https: true },
	optimizeDeps: {
		exclude: ["@jsquash/jpeg", "@jsquash/png", "@jsquash/resize"]
	}
});

// Authentication handling

const Auth = {
    async checkRequired() {
        try {
            const resp = await fetch('/api/auth/status');
            const data = await resp.json();
            return data.required;
        } catch {
            return false;
        }
    },

    async login(pin, remember) {
        const resp = await fetch('/api/auth', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ pin, remember }),
        });

        const data = await resp.json();

        if (!resp.ok) {
            throw new Error(data.error || 'Authentication failed');
        }

        return data;
    },

    setupLoginForm() {
        const form = document.getElementById('login-form');
        const error = document.getElementById('login-error');

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            error.style.display = 'none';

            const pin = document.getElementById('pin-input').value;
            const remember = document.getElementById('remember-me').checked;

            try {
                await Auth.login(pin, remember);
                document.getElementById('login-screen').style.display = 'none';
                document.querySelector('.app').classList.add('active');
                App.init();
            } catch (err) {
                error.textContent = err.message;
                error.style.display = 'block';
            }
        });
    }
};

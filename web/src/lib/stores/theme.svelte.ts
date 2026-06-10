type ThemePreference = 'light' | 'dark' | 'system';

function createThemeStore() {
  const stored = typeof localStorage !== 'undefined'
    ? localStorage.getItem('theme-preference') as ThemePreference | null
    : null;

  let _preference = $state<ThemePreference>(stored ?? 'system');

  function resolve(pref: ThemePreference): 'light' | 'dark' {
    if (pref === 'system') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return pref;
  }

  let _effective = $state<'light' | 'dark'>(resolve(_preference));

  let detach: (() => void) | null = null;

  function attachListener() {
    detach?.();
    detach = null;
    if (_preference === 'system') {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      const handler = () => { _effective = resolve('system'); };
      mq.addEventListener('change', handler);
      detach = () => mq.removeEventListener('change', handler);
    }
  }

  attachListener();

  return {
    get preference() { return _preference; },
    set preference(v: ThemePreference) {
      _preference = v;
      localStorage.setItem('theme-preference', v);
      _effective = resolve(v);
      attachListener();
    },
    get effective() { return _effective; },
  };
}

export const theme = createThemeStore();

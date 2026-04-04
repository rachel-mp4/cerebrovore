import { mount } from 'svelte'
import Settings from './Settings.svelte'

const app = mount(Settings, {
  target: document.getElementById('settings-pane')!,
})

export default app

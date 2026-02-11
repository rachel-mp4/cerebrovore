import { mount } from 'svelte'
import Beep from './Beep.svelte'

const app = mount(Beep, {
  target: document.getElementById('cbv')!,
})

export default app

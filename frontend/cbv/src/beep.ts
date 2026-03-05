import { mount } from 'svelte'
import Beep from './Beep.svelte'

const app = mount(Beep, {
  target: document.getElementById('0cbv')!,
})

export default app

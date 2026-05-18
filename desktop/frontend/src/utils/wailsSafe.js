import { EventsOff, EventsOn } from '../../wailsjs/runtime'

export async function safeInvoke(factory, fallbackValue, label = 'Wails call unavailable') {
  try {
    return await factory()
  } catch (error) {
    if (import.meta.env.DEV) {
      console.debug(label, error)
    }
    return typeof fallbackValue === 'function' ? fallbackValue(error) : fallbackValue
  }
}

export function safeEventsOn(name, handler) {
  try {
    EventsOn(name, handler)
  } catch (error) {
    if (import.meta.env.DEV) {
      console.debug(`Wails runtime event unavailable: ${name}`, error)
    }
  }
}

export function safeEventsOff(name) {
  try {
    EventsOff(name)
  } catch (error) {
    if (import.meta.env.DEV) {
      console.debug(`Wails runtime event unavailable: ${name}`, error)
    }
  }
}

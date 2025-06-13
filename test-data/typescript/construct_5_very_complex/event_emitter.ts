// Utility type to get event names with a 'Changed' suffix.
export type ChangeEvent<T extends string> = `${T}Changed`;

// Utility type to infer the payload from a function signature.
type Payload<T> = T extends (payload: infer P) => void ? P : never;

// Conditional type to map event names to their listener functions.
export type ListenerMap<TEventMap extends object> = {
  [K in keyof TEventMap as ChangeEvent<K & string>]: (payload: TEventMap[K]) => void;
};

export class TypedEventEmitter<TEventMap extends object> {
  private listeners: Partial<ListenerMap<TEventMap>> = {};

  // The 'on' method's signature is dynamically typed based on TEventMap.
  public on<TEventName extends keyof ListenerMap<TEventMap>>(
    eventName: TEventName,
    listener: ListenerMap<TEventMap>[TEventName]
  ): void {
    this.listeners[eventName] = listener;
  }

  // The 'emit' method's signature is also dynamically typed.
  public emit<TEventName extends keyof ListenerMap<TEventMap>>(
    eventName: TEventName,
    payload: Payload<ListenerMap<TEventMap>[TEventName]>
  ): void {
    const listener = this.listeners[eventName];
    if (listener) {
      // We need a type assertion here due to TS limitations with complex indexed access.
      (listener as (p: typeof payload) => void)(payload);
    }
  }
}

// Example Usage (The distiller should be able to handle this too)
interface AppEvents {
  user: { id: number; name: string };
  settings: { theme: 'dark' | 'light' };
}

const appEmitter = new TypedEventEmitter<AppEvents>();
appEmitter.on('userChanged', (payload) => console.log(payload.id));
// appEmitter.on('settingsChanged', (payload) => console.log(payload.id)); // This would be a type error
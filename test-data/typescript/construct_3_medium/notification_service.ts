// Interface for any object that can receive a notification payload.
export interface INotifiable<P> {
  handleNotification(payload: P): void;
}

// Generic payload structure
export type NotificationPayload<T> = {
  recipient: T;
  message: string;
  timestamp: Date;
};

abstract class BaseNotificationService {
  protected transportName: string;

  constructor(transportName: string) {
    this.transportName = transportName;
  }

  // Abstract method must be implemented by subclasses.
  public abstract send<T extends { id: string | number }>(payload: NotificationPayload<T>): Promise<boolean>;

  protected log(message: string): void {
    console.log(`[${this.transportName}] ${message}`);
  }
}

export class EmailNotificationService extends BaseNotificationService {
  constructor() {
    super("EMAIL");
  }

  public async send<T extends { id: string | number; email: string }>(payload: NotificationPayload<T>): Promise<boolean> {
    this.log(`Preparing to send email to user ${payload.recipient.id}.`);
    // Fake sending logic
    await new Promise(resolve => setTimeout(resolve, 100));
    this.log(`Email sent to ${payload.recipient.email}.`);
    return true;
  }
}
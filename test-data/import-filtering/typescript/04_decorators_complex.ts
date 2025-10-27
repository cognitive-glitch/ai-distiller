// Test Pattern 4: Decorators, Conditional Imports, and Complex Type Imports
// Tests decorator imports, environment-based imports, and advanced type patterns

import { Injectable, Inject } from '@angular/core';
import { Observable, Subject, BehaviorSubject } from 'rxjs';
import { map, filter, tap, catchError } from 'rxjs/operators';
import type { User, Admin, Permission } from './models';
import { environment } from '../environments/environment';
import { BaseService } from './base.service';
import 'reflect-metadata'; // Side-effect for decorators

// Conditional imports based on environment
if (environment.production) {
  import('@sentry/browser').then(Sentry => {
    Sentry.init({ dsn: environment.sentryDsn });
  });
} else {
  import('redux-devtools-extension').then(({ composeWithDevTools }) => {
    window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ = composeWithDevTools;
  });
}

// Import used only in decorator metadata
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { Validator } from './validators';

// Not using Subject, tap, Admin, BaseService, or Validator

// Decorator usage
@Injectable({
  providedIn: 'root'
})
export class UserService {
  private users$ = new BehaviorSubject<User[]>([]);

  constructor(
    @Inject(HTTP_INTERCEPTORS) private interceptors: any[]
  ) {}

  getUsers(): Observable<User[]> {
    return this.users$.asObservable().pipe(
      map(users => users.filter(u => u.active)),
      filter(users => users.length > 0),
      catchError(error => {
        console.error('Error in getUsers:', error);
        return [];
      })
    );
  }

  checkPermission(user: User, permission: Permission): boolean {
    return user.permissions.includes(permission);
  }

  // Method using environment
  getApiUrl(): string {
    return environment.apiUrl;
  }
}

// Type used in conditional type
type UserOrAdmin<T> = T extends { isAdmin: true } ? Admin : User;

// Complex type intersection
type AuthorizedUser = User & { token: string; permissions: Permission[] };
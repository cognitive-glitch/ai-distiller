// 02_simple.cs
// Simple domain model with encapsulation, XML-docs, events and struct record.

#nullable enable
using System;

namespace Constructs.Simple02;

/// <summary>
/// Represents a bank account with rudimentary business logic.
/// </summary>
public class BankAccount
{
    private decimal _balance;

    /// <summary>Raised whenever money is deposited or withdrawn.</summary>
    public event EventHandler<decimal>? BalanceChanged;

    /// <summary>Unique, read-only account number.</summary>
    public string AccountNumber { get; }

    /// <summary>Current balance (read-only outside the class).</summary>
    public decimal Balance
    {
        get => _balance;
        protected set
        {
            _balance = value;
            BalanceChanged?.Invoke(this, _balance);
        }
    }

    public BankAccount(string accountNumber, decimal openingBalance = 0m)
        => (AccountNumber, Balance) = (accountNumber, openingBalance);

    /// <exception cref="ArgumentOutOfRangeException"/>
    public void Deposit(decimal amount)
    {
        if (amount <= 0) throw new ArgumentOutOfRangeException(nameof(amount));
        Balance += amount;
    }

    /// <exception cref="InvalidOperationException"/>
    public void Withdraw(decimal amount)
    {
        if (amount > Balance) throw new InvalidOperationException("Insufficient funds.");
        Balance -= amount;
    }

    /// <summary>
    /// Private method for audit logging
    /// </summary>
    private void LogTransaction(string type, decimal amount)
    {
        Console.WriteLine($"[{DateTime.Now}] {type}: {amount:C} on account {AccountNumber}");
    }

    /// <summary>
    /// Protected method for validation
    /// </summary>
    protected virtual bool ValidateTransaction(decimal amount)
    {
        return amount > 0 && amount <= 10000; // Max transaction limit
    }

    /// <summary>
    /// Internal method for bank operations
    /// </summary>
    internal void ProcessInterest(decimal rate)
    {
        if (ValidateTransaction(Balance * rate))
        {
            Balance += Balance * rate;
            LogTransaction("Interest", Balance * rate);
        }
    }

    public override string ToString() => $"{AccountNumber}: {Balance:C}";
}

/// <summary>
/// Immutable value object using readonly struct.
/// </summary>
public readonly struct Money : IEquatable<Money>
{
    public Money(decimal amount, string currency)
        => (Amount, Currency) = (amount, currency);

    public decimal Amount   { get; }
    public string  Currency { get; }

    public bool Equals(Money other)
        => Amount == other.Amount && Currency == other.Currency;

    public override int GetHashCode() => HashCode.Combine(Amount, Currency);

    public override string ToString() => $"{Amount:0.00} {Currency}";

    /// <summary>
    /// Implicit conversion from decimal
    /// </summary>
    public static implicit operator Money(decimal amount)
        => new Money(amount, "USD");

    /// <summary>
    /// Addition operator for money
    /// </summary>
    public static Money operator +(Money left, Money right)
    {
        if (left.Currency != right.Currency)
            throw new InvalidOperationException("Cannot add different currencies");
        
        return new Money(left.Amount + right.Amount, left.Currency);
    }
}

/// <summary>
/// Savings account with interest calculation
/// </summary>
public class SavingsAccount : BankAccount
{
    private decimal _interestRate;

    /// <summary>
    /// Interest rate as percentage
    /// </summary>
    public decimal InterestRate 
    { 
        get => _interestRate;
        set => _interestRate = Math.Max(0, Math.Min(value, 10)); // 0-10% range
    }

    public SavingsAccount(string accountNumber, decimal interestRate, decimal openingBalance = 0m) 
        : base(accountNumber, openingBalance)
    {
        InterestRate = interestRate;
    }

    /// <summary>
    /// Override validation for savings accounts
    /// </summary>
    protected override bool ValidateTransaction(decimal amount)
    {
        return base.ValidateTransaction(amount) && Balance - amount >= 100; // Minimum balance
    }

    /// <summary>
    /// Calculate and add monthly interest
    /// </summary>
    public void AddMonthlyInterest()
    {
        ProcessInterest(InterestRate / 12 / 100);
    }
}
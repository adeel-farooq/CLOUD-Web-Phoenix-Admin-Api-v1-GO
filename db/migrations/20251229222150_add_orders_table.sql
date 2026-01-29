-- Create ApprovalTypes table
CREATE TABLE [dbo].[ApprovalTypes] (
    [Id] INT PRIMARY KEY IDENTITY(1,1),
    [Name] NVARCHAR(100) NOT NULL,
    [Description] NVARCHAR(500),
    [RequiresMultipleApprovals] BIT DEFAULT 0,
    [ApprovalLevels] INT DEFAULT 1,
    [MaxApprovalTimeMinutes] INT DEFAULT 1440,
    [IsActive] BIT DEFAULT 1,
    [CreatedDate] DATETIME DEFAULT GETUTCDATE(),
    [CreatedBy] NVARCHAR(100),
    [ModifiedDate] DATETIME,
    [ModifiedBy] NVARCHAR(100)
);

-- Create Approvals table
CREATE TABLE [dbo].[Approvals] (
    [Id] INT PRIMARY KEY IDENTITY(1,1),
    [ApprovalTypeId] INT NOT NULL,
    [ReferenceId] NVARCHAR(100) NOT NULL,
    [ReferenceType] NVARCHAR(50),
    [Amount] DECIMAL(18, 2),
    [Status] NVARCHAR(50) DEFAULT 'PENDING',
    [CreatedDate] DATETIME DEFAULT GETUTCDATE(),
    [CreatedBy] NVARCHAR(100),
    [Notes] NVARCHAR(MAX),
    [ApprovedDate] DATETIME,
    [ApprovedBy] NVARCHAR(100),
    [ApprovalNotes] NVARCHAR(MAX),
    FOREIGN KEY ([ApprovalTypeId]) REFERENCES [ApprovalTypes]([Id])
);

-- Insert sample data - ADD CreatedBy column
INSERT INTO [ApprovalTypes] (Name, Description, RequiresMultipleApprovals, ApprovalLevels, MaxApprovalTimeMinutes, IsActive, CreatedBy)
VALUES 
    ('Transfer Approval', 'Transfers exceeding 500,000 PKR require approval', 1, 2, 1440, 1, 'system'),
    ('Business Onboarding', 'New business registration requires approval', 1, 3, 2880, 1, 'system'),
    ('High Risk Transaction', 'Transactions flagged as high risk', 0, 1, 480, 1, 'system'),
    ('Refund Approval', 'Refunds exceeding 100,000 PKR', 1, 2, 1440, 1, 'system'),
    ('Compliance Review', 'KYC/AML compliance check required', 1, 2, 2880, 1, 'system');

-- Create indexes
CREATE INDEX [IDX_ApprovalTypes_IsActive] ON [ApprovalTypes]([IsActive]);
CREATE INDEX [IDX_Approvals_ApprovalTypeId] ON [Approvals]([ApprovalTypeId]);
CREATE INDEX [IDX_Approvals_Status] ON [Approvals]([Status]);


ALTER TABLE ApprovalTypes
ADD ModifiedBy NVARCHAR(100) NULL;

ALTER TABLE ApprovalTypes
ADD ModifiedDate DATETIME NULL;

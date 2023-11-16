-- ref: https://en.wikipedia.org/wiki/List_of_circulating_currencies
-- design decision - store all amounts in fractional units
-- application keeps as constants numbers to basic for currencies
CREATE TABLE CURRENCY(
    ID CHAR(3) PRIMARY KEY, -- iso
    DESCRIPTION VARCHAR(200),
    FRACTIONAL_UNIT VARCHAR(20)
);

CREATE TABLE OPERATION_TYPE(
    ID VARCHAR(3) PRIMARY KEY,
    DESCRIPTION VARCHAR(200)
);

CREATE TABLE OPERATION_STATUS(
    ID VARCHAR(10) PRIMARY KEY,
    DESCRIPTION VARCHAR(200)
);

-- subsystems that ledger system should know
CREATE TABLE SUBSYSTEM(
     ID UUID PRIMARY KEY, -- unique identifier
     NAME VARCHAR(10),
     DESCRIPTION VARCHAR(200)
);

-- main table to store ACCOUNTS
CREATE TABLE ACCOUNT(
    ID UUID PRIMARY KEY, -- unique identifier
    ALLOW_ASYNC BOOLEAN NOT NULL, -- if we can withdraw money asynchronously
    IS_SUSPENDED BOOLEAN NOT NULL, -- account can be suspended for a reason
    SUBSYSTEM_ID UUID NOT NULL, -- subsystem, that created account
    CREATED_DATE TIMESTAMPTZ NOT NULL,
    SUSPENDED_DATE TIMESTAMPTZ,
    SUSPENDED_REASON VARCHAR(200),
    FOREIGN KEY (SUBSYSTEM_ID) REFERENCES SUBSYSTEM(ID)
);

CREATE TABLE ACCOUNT_CURRENCY(
    ID UUID PRIMARY KEY,
    ACCOUNT_ID UUID NOT NULL,
    CURRENCY_ID CHAR(3) NOT NULL,
    AMOUNT BIGINT NOT NULL,
    CREDIT_LIMIT BIGINT, -- IF NULL - THEN NO LIMIT
    FOREIGN KEY (ACCOUNT_ID) REFERENCES ACCOUNT(ID),
    FOREIGN KEY (CURRENCY_ID) REFERENCES CURRENCY(ID)
);

CREATE TABLE TRANSACTION(
    ID UUID PRIMARY KEY,
    SUBSYSTEM_ID UUID NOT NULL,
    SOURCE_ACCOUNT_CURRENCY_ID UUID NOT NULL,
    TARGET_ACCOUNT_CURRENCY_ID UUID NOT NULL,
    TRANSACTION_DATE TIMESTAMPTZ NOT NULL,
    AMOUNT BIGINT NOT NULL,
    CURRENCY_ID CHAR(3) NOT NULL,
    DESCRIPTION VARCHAR(200) NOT NULL,
    FOREIGN KEY (SUBSYSTEM_ID) REFERENCES SUBSYSTEM(ID),
    FOREIGN KEY (CURRENCY_ID) REFERENCES CURRENCY(ID),
    FOREIGN KEY (SOURCE_ACCOUNT_CURRENCY_ID) REFERENCES ACCOUNT_CURRENCY(ID),
    FOREIGN KEY (TARGET_ACCOUNT_CURRENCY_ID) REFERENCES ACCOUNT_CURRENCY(ID)
);

CREATE TABLE ACCOUNT_OPERATION(
    ID BIGSERIAL PRIMARY KEY,
    ACCOUNT_CURRENCY_ID UUID NOT NULL,
    OPERATION_TYPE_ID VARCHAR(3) NOT NULL,
    OPERATION_STATUS_ID VARCHAR(10) NOT NULL,
    AMOUNT BIGINT NOT NULL,
    TRANSACTION_ID UUID NOT NULL,
    FOREIGN KEY (ACCOUNT_CURRENCY_ID) REFERENCES ACCOUNT_CURRENCY(ID),
    FOREIGN KEY (OPERATION_TYPE_ID) REFERENCES OPERATION_TYPE(ID),
    FOREIGN KEY (TRANSACTION_ID) REFERENCES TRANSACTION(ID),
    FOREIGN KEY (OPERATION_STATUS_ID) REFERENCES OPERATION_STATUS(ID)
);

-- table is required to store operations that need to be processed
CREATE TABLE ACCOUNT_OPERATION_OUTPOST(
    ID BIGINT PRIMARY KEY,
    CREATED_DATE TIMESTAMPTZ NOT NULL
);
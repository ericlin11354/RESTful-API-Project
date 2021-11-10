DROP TABLE IF EXISTS TimeSeries CASCADE;
DROP TABLE IF EXISTS TimeSeriesDate CASCADE;

CREATE TABLE TimeSeries(
	ID INT AUTO_INCREMENT,
	Admin2 TEXT,
	Address1 TEXT,
	Address2 TEXT,
	primary key(ID)
);

CREATE TABLE TimeSeriesDate(
	ID INT,
	Date Date,
	Confirmed INT,
	Death INT,
	Recovered INT,
	Primary Key (ID, Date),
	FOREIGN KEY (ID) REFERENCES TimeSeries(ID)
);

CREATE TABLE DailyReports(
	ID INT AUTO_INCREMENT,
	Date Date NOT NULL,
	Admin2 TEXT,
	Address1 TEXT,
	Address2 TEXT,
	Confirmed INT,
	Death INT,
	Recovered INT,
	Active INT,
	primary key(ID)
);
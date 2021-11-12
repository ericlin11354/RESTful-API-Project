DROP TABLE IF EXISTS TimeSeriesConfirmed CASCADE;
DROP TABLE IF EXISTS TimeSeriesDeath CASCADE;
DROP TABLE IF EXISTS TimeSeriesRecovered CASCADE;
DROP TABLE IF EXISTS TimeSeries CASCADE;
DROP TABLE IF EXISTS DailyReports CASCADE;

CREATE TABLE TimeSeries(
	ID INT AUTO_INCREMENT,
	Admin2 VARCHAR(128),
	Address1 VARCHAR(128),
	Address2 VARCHAR(128) NOT NULL,
	PRIMARY KEY(ID),
	CONSTRAINT AddressKey UNIQUE (Admin2,Address1,Address2)
);

CREATE TABLE TimeSeriesConfirmed(
	ID INT,
	Date Date NOT NULL,
	Confirmed INT NOT NULL,
	PRIMARY KEY(ID, Date),
	FOREIGN KEY (ID) REFERENCES TimeSeries(ID)
);

CREATE TABLE TimeSeriesDeath(
	ID INT,
	Date Date NOT NULL,
	Death INT NOT NULL,
	PRIMARY KEY(ID, Date),
	FOREIGN KEY (ID) REFERENCES TimeSeries(ID)
);

CREATE TABLE TimeSeriesRecovered(
	ID INT,
	Date Date NOT NULL,
	Recovered INT NOT NULL,
	PRIMARY KEY(ID, Date),
	FOREIGN KEY (ID) REFERENCES TimeSeries(ID)
);

INSERT INTO TimeSeries(Admin2, Address1, Address2)
VALUES('Autauga', 'Alabama', 'US');

INSERT INTO TimeSeries(Address1, Address2)
VALUES('Ontario', 'Canada');

INSERT INTO TimeSeriesConfirmed VALUES
(1, "2020/01/31", 10),
(1, "2021/11/1", 420),
(2, "2020/01/31", 1),
(2, "2021/10/31", 343);

INSERT INTO TimeSeriesDeath VALUES
(1, "2020/01/31", 20),
(1, "2021/11/1", 69),
(2, "2020/01/31", 2),
(2, "2021/10/31", 369);

INSERT INTO TimeSeriesRecovered VALUES
(1, "2020/01/31", 30),
(1, "2021/11/1", 301),
(2, "2020/01/31", 3),
(2, "2021/10/31", 311);

CREATE TABLE DailyReports(
	ID INT AUTO_INCREMENT,
	Date Date NOT NULL,
	Admin2 VARCHAR(128),
	Address1 VARCHAR(128),
	Address2 VARCHAR(128) NOT NULL,
	Confirmed INT,
	Death INT,
	Recovered INT,
	Active INT,
	PRIMARY KEY(ID),
	CONSTRAINT ADKey UNIQUE (Date,Admin2,Address1,Address2)
);

INSERT INTO DailyReports(Date, Admin2, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES
("2020/06/05", 'Abbeville', 'South Carolina', 'US', 47,0,0,47),
("2020/01/31", 'Abbeville', 'South Carolina', 'US', 1,2,3,4);
INSERT INTO DailyReports(Date, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES
("2020/02/14", 'Ontario', 'Canada', 5,6,7,8),
("2020/11/16", 'British Columbia', 'Canada', 301,343,369,373);
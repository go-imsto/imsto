ALTER TABLE map_template ADD roofs varCHAR(12)[] NOT NULL DEFAULT '{}';
ALTER TABLE map_template ADD created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;

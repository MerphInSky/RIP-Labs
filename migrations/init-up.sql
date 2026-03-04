CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    is_moderator BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS components (
    component_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    photo_url VARCHAR(255),
    video VARCHAR(255),
    thermal_resistance NUMERIC NOT NULL DEFAULT 1.0
);

CREATE TABLE IF NOT EXISTS heatings (
    heating_id SERIAL PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    creator_id INTEGER NOT NULL REFERENCES users(user_id),
    forming_date TIMESTAMP,
    finish_date TIMESTAMP,
    moderator_id INTEGER REFERENCES users(user_id),
    ambient_temperature NUMERIC NOT NULL DEFAULT 1.0,
    description VARCHAR(2000)
);

CREATE TABLE IF NOT EXISTS heating_components (
    heating_id INTEGER NOT NULL REFERENCES heatings(heating_id),
    component_id INTEGER NOT NULL REFERENCES components(component_id),
    power_dissipation NUMERIC NOT NULL DEFAULT 5.0,
    heat NUMERIC,
    PRIMARY KEY (heating_id, component_id)
);

INSERT INTO users (login, password, is_moderator) VALUES
    ('user1', 'pass1', false),
    ('moderator', 'modpass', true)
ON CONFLICT (login) DO NOTHING;

INSERT INTO components (title, description, is_deleted, photo_url, video, thermal_resistance) VALUES
    ('Микросхема BD9897FS', 'Контроллер управления инвертором подсветки для ЖК-дисплеев в корпусе для поверхностного монтажа SOP-24', false, 'Num1.jpg', 'gifgif.mp4', 132.0),
    ('Транзистор FDD8447L', 'N-канальный полевой транзистор для силовой коммутации в корпусе D-PAK для поверхностного монтажа', false, 'Num2.jpg', 'GIFKA.mp4', 2.3),
    ('Транзистор B1261', 'Высокотоковый PNP-транзистор для усилителей мощности и схем управления в металлическом корпусе TO-3P', false, 'Num3.jpg', 'gifgif.mp4', 12.5),
    ('Транзисторная сборка AO4606C', 'Комплементарная пара полевых транзисторов N- и P-канального типов в компактном корпусе SO-8 для поверхностного монтажа', false, 'Num4.jpg', 'GIFKA.mp4', 40.0),
    ('Стабилитрон BZX55-33V', 'Малосигнальный стабилитрон, напряжение стабилизации 33 В, мощность 500 мВт, допуск ±5%, корпус DO-35 (выводной монтаж) ', false, 'Num5.jpg', 'gifgif.mp4', 300.0),
    ('Микросхема NE55H', 'Контроллер импульсного источника питания с ШИМ, токовый режим управления, защита от перегрузки, корпус DIP-8', false, 'Num6.jpg', 'GIFKA.mp4', 15.0);

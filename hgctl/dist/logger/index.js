"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupLogger = setupLogger;
const winston_1 = __importDefault(require("winston"));
const chalk_1 = __importDefault(require("chalk"));
class WinstonLogger {
    winston;
    constructor(verbose = false) {
        const format = winston_1.default.format.combine(winston_1.default.format.timestamp({ format: 'HH:mm:ss' }), winston_1.default.format.printf(({ level, message, timestamp }) => {
            const color = {
                info: chalk_1.default.blue,
                warn: chalk_1.default.yellow,
                error: chalk_1.default.red,
                debug: chalk_1.default.gray
            }[level] || chalk_1.default.white;
            return color(`[${timestamp}] ${level.toUpperCase()}: ${message}`);
        }));
        this.winston = winston_1.default.createLogger({
            level: verbose ? 'debug' : 'info',
            format,
            transports: [
                new winston_1.default.transports.Console()
            ]
        });
    }
    info(message, ...args) {
        this.winston.info(this.format(message, args));
    }
    warn(message, ...args) {
        this.winston.warn(this.format(message, args));
    }
    error(message, ...args) {
        this.winston.error(this.format(message, args));
    }
    debug(message, ...args) {
        this.winston.debug(this.format(message, args));
    }
    title(message, ...args) {
        console.log(chalk_1.default.bold.cyan(`\n${this.format(message, args)}\n`));
    }
    format(message, args) {
        return args.length > 0 ? message.replace(/%s/g, () => args.shift()) : message;
    }
}
function setupLogger(verbose = false) {
    return new WinstonLogger(verbose);
}
//# sourceMappingURL=index.js.map
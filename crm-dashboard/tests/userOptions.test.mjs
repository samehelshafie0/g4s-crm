import test from 'node:test'
import assert from 'node:assert/strict'
import {passwordError} from '../src/services/userOptions.ts'
test('password forms reject mismatches and bcrypt byte overflows', () => {
 assert.match(passwordError('short','short'), /at least 8/)
 assert.match(passwordError('longEnough1','different'), /do not match/)
 const unicode = '🔒'.repeat(20)
 assert.match(passwordError(unicode,unicode), /72 bytes/)
 assert.equal(passwordError('A valid password 42','A valid password 42'),'')
})

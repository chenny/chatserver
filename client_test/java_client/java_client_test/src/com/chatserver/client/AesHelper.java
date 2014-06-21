package com.chatserver.client;

import javax.crypto.Cipher;
import javax.crypto.spec.SecretKeySpec;
import javax.crypto.spec.IvParameterSpec;

public class AesHelper {
	private IvParameterSpec ivSpec;
	private SecretKeySpec keySpec;
	
	public AesHelper(String key) {
		try {
			this.keySpec = new SecretKeySpec(key.getBytes(), "AES");
			this.ivSpec = new IvParameterSpec(key.getBytes());
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

	public byte[] encrypt(byte[] origData) {
		try {
			Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
			cipher.init(Cipher.ENCRYPT_MODE, this.keySpec, this.ivSpec);
			return cipher.doFinal(origData);
		} catch (Exception e) {
			e.printStackTrace();
		}
		return null;
	}

	public byte[] decrypt(byte[] crypted) {
		try {
			Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
			cipher.init(Cipher.DECRYPT_MODE, this.keySpec, this.ivSpec);
			return cipher.doFinal(crypted);
		} catch (Exception e) {
			e.printStackTrace();
		}
		return null;
	}
	
	public static void main(String[] args) {
		AesHelper aes = new AesHelper("12345abcdef67890");
		String data = "hello world我爱你， 今晚嘿咻，这是一个测试";
		byte[] encrypted = aes.encrypt(data.getBytes());
		byte[] decrypted = aes.decrypt(encrypted);
		System.out.println(new String(decrypted));
	}
	
}

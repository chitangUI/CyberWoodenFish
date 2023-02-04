using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

namespace ElectronicFish
{

	public class WelcomeManager : MonoBehaviour
	{
		[SerializeField] private Text comboText;
		[SerializeField] private Button startGameButton;

		private void Start()
		{
			var combo = PlayerPrefs.GetInt("combo");

			startGameButton.onClick.AddListener(StartGame);
			startGameButton.onClick.AddListener(delegate { Debug.Log("nmsl"); });
			comboText.text = $"您累计敲了 {combo} 次";

		}

		private static void StartGame()
		{
			Debug.Log("nmsl");

			SceneManager.LoadScene("Scenes/MainScene");
		}
	}

}
